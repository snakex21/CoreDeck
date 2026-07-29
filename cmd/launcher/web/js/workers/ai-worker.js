// ============================================
// AI Worker — LLM generative search only
// Uses local llama-server.exe via OpenAI-compatible API.
// No ONNX / transformers.js / embedding models.
// ============================================

let llmServer = '';

let programs = [];

function buildCatalogSummary(progs) {
  return progs.map((p, index) => ({
    id: index,
    name: p.name,
    desc: p.description || '',
    cat: p.category || '',
    tags: (p.tags || []).join(', '),
    path: p.folderPath || p.localPath || '',
    type: p.type || 'folder',
  }));
}

function rankCandidates(query, progs, limit = 40) {
  const q = (query || '').toLowerCase().trim();
  const tokens = q.split(/\s+/).filter(Boolean);
  if (!q) return progs.slice(0, limit);

  const qWords = q.match(/[\p{L}\p{N}]+/gu) || [];
  const isShortQuery = q.length <= 4 && qWords.length === 1;

  function wordsOf(text) {
    return ((text || '').toLowerCase().match(/[\p{L}\p{N}]+/gu) || []);
  }

  function commonPrefixLength(a, b) {
    let i = 0;
    const max = Math.min(a.length, b.length);
    while (i < max && a[i] === b[i]) i++;
    return i;
  }

  function relatedLongWord(token, word) {
    if (!token || !word) return false;
    if (token.length < 6 || word.length < 6) return false;
    const common = commonPrefixLength(token, word);
    const required = Math.max(5, Math.min(token.length, word.length) - 2);
    return common >= required;
  }

  function fieldScore(text, weight) {
    const lower = (text || '').toLowerCase();
    const words = wordsOf(lower);
    let score = 0;

    if (lower === q) score += weight * 4;
    if (words.includes(q)) score += weight * 3;
    if (!isShortQuery && lower.includes(q)) score += weight * 1.4;

      for (const token of tokens) {
        if (words.includes(token)) score += weight * 1.5;
        else if (!isShortQuery && words.some(w => relatedLongWord(token, w))) score += weight * 1.05;
        else if (!isShortQuery && words.some(w => w.startsWith(token))) score += weight * 0.8;
        else if (!isShortQuery && lower.includes(token)) score += weight * 0.3;
      }

    return score;
  }

  return progs
    .map(program => {
      let score = 0;
      score += fieldScore(program.name, 80);
      score += fieldScore((program.tags || []).join(' '), 45);
      score += fieldScore(program.category, 30);
      score += fieldScore(program.folderPath || program.localPath || '', 28);
      score += fieldScore(program.description, isShortQuery ? 8 : 18);

      if (isShortQuery && score > 0) {
        const nameWords = wordsOf(program.name);
        const tagWords = wordsOf((program.tags || []).join(' '));
        const catWords = wordsOf(program.category);
        const strongHit = nameWords.includes(q) || tagWords.includes(q) || catWords.includes(q);
        if (!strongHit) score *= 0.18;
      }

      return { program, score };
    })
    .filter(r => r.score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
    .map(r => r.program);
}

function normalizeName(name) {
  return (name || '').toLowerCase().trim();
}

function findProgramFromCandidate(candidate, catalog, candidatePrograms) {
  if (!candidate || !candidatePrograms.length) return null;

  const id = Number(candidate.id);
  if (Number.isInteger(id) && id >= 0 && id < candidatePrograms.length) {
    return candidatePrograms[id];
  }

  const wanted = normalizeName(candidate.name);
  if (!wanted) return null;

  const catalogItem = catalog.find(item => normalizeName(item.name) === wanted);
  if (catalogItem && Number.isInteger(catalogItem.id)) {
    return candidatePrograms[catalogItem.id] || null;
  }

  return null;
}

async function llmSearch(query, allPrograms, maxResults = 10, requestId = 0) {
  const candidates = rankCandidates(query, allPrograms, 40);

  if (!candidates.length) {
    postMessage({ type: 'result', requestId, query, results: [] });
    return;
  }

  const catalog = buildCatalogSummary(candidates);

  const systemPrompt = `Masz dostępny katalog programów i narzędzi. Użytkownik opisuje czego szuka.
To jest nowe, niezależne wyszukiwanie. Ignoruj jakiekolwiek poprzednie zapytania i odpowiedzi.
Zwróć dokładnie listę JSON z pasującymi programami.
Odpowiadaj TYLKO i wyłącznie poprawnym JSON — żadnych komentarzy, markdownu ani tekstu przed/po.
Pisz po polsku z poprawnymi polskimi znakami diakrytycznymi (ą, ć, ę, ł, ń, ó, ś, ź, ż).

Zasady dopasowania:
- Krótkie zapytania traktuj jako konkretne słowa kluczowe, nie zgaduj podobnych słów.
- Przykład: "audi" oznacza markę Audi/VAG/auto, NIE "audio" ani programy dźwiękowe.
- Preferuj trafienia w nazwie, tagach i kategorii. Opis jest tylko pomocniczy.
- Nie zwracaj pozycji luźno podobnych, jeśli nie są realnie związane z intencją.

Format odpowiedzi — zwracaj tylko id z katalogu:
[
  {"id": 0, "name": "nazwa programu", "reason": "krótki powód"},
  {"id": 3, "name": "nazwa programu 2", "reason": "krótki powód"}
]

Jeśli nic nie pasuje, zwróć pustą listę: []

Katalog programów:
${JSON.stringify(catalog, null, 0)}`;

  const messages = [
    { role: 'system', content: systemPrompt },
    { role: 'user', content: query },
  ];

  postMessage({ type: 'status', status: 'loading_model', model: 'llm', modelName: 'lfm2.5-350m' });

  try {
    if (!llmServer) throw new Error('Lokalny serwer AI nie został uruchomiony');
    const resp = await fetch(`${llmServer}/v1/chat/completions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages,
        temperature: 0,
        max_tokens: 512,
        stream: false,
      }),
    });

    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}));
      throw new Error(err.error?.message || `Server ${resp.status}`);
    }

    const data = await resp.json();
    postMessage({ type: 'status', status: 'model_loaded', model: 'llm' });

    const content = data.choices?.[0]?.message?.content || '';

    // Parse JSON from LLM response (may be wrapped in markdown code block)
    let raw = (content || '').trim();
    const fenceMatch = raw.match(/```(?:json)?\s*([\s\S]*?)```/);
    if (fenceMatch) raw = fenceMatch[1].trim();

    let matches = [];
    try {
      matches = JSON.parse(raw);
    } catch {
      const arrMatch = raw.match(/\[[\s\S]*\]/);
      if (arrMatch) {
        try { matches = JSON.parse(arrMatch[0]); } catch { matches = []; }
      }
    }

    if (!Array.isArray(matches)) matches = [];

    const results = matches
      .slice(0, maxResults)
      .map(m => {
        const program = findProgramFromCandidate(m, catalog, candidates);
        return program ? { program, similarity: 0.9, reason: m.reason || '' } : null;
      })
      .filter(Boolean);

    if (!results.length) {
      const fallback = (candidates.length ? candidates : allPrograms)
        .slice(0, maxResults)
        .map(program => ({ program, similarity: 0.5, reason: 'Dopasowanie lokalne' }));
      postMessage({ type: 'result', requestId, query, results: fallback });
      return;
    }

    postMessage({ type: 'result', requestId, query, results });
  } catch (error) {
    const fallback = rankCandidates(query, allPrograms, maxResults)
      .map(program => ({ program, similarity: 0.4, reason: 'Fallback lokalny po błędzie LLM' }));
    if (fallback.length) {
      postMessage({ type: 'result', requestId, query, results: fallback });
    } else {
      postMessage({ type: 'error', requestId, error: `LLM: ${error.message}` });
    }
  }
}

// ============================================
// Message handler
// ============================================
self.onmessage = async (event) => {
  const msg = event.data || {};
  try {
    if (msg.type === 'init') {
      programs = msg.programs || [];
      llmServer = msg.serverUrl || '';
      postMessage({ type: 'ready', model: 'llm', total: programs.length });
    } else if (msg.type === 'search') {
      await llmSearch(msg.query || '', programs, msg.maxResults || 10, msg.requestId || 0);
    }
  } catch (error) {
    postMessage({ type: 'error', error: error?.message || String(error) });
  }
};
