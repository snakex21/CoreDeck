let llmServer = '';

let links = [];

function normalized(value) {
  return String(value || '')
    .toLocaleLowerCase('pl')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '');
}

function words(value) {
  return normalized(value).match(/[\p{L}\p{N}]+/gu) || [];
}

function rankLinks(question, limit = 24) {
  const tokens = words(question).filter(token => token.length > 1);
  if (!tokens.length) return links.slice(0, limit);

  return links.map(link => {
    const fields = [
      [link.name, 16], [link.tag, 10], [link.section, 9],
      [link.category, 8], [link.group, 6], [link.note, 5],
    ];
    let score = 0;
    for (const [value, weight] of fields) {
      const text = normalized(value);
      const fieldWords = words(value);
      for (const token of tokens) {
        if (fieldWords.includes(token)) score += weight * 3;
        else if (text.includes(token)) score += weight;
        else if (token.length >= 5 && fieldWords.some(word => word.startsWith(token.slice(0, -1)))) score += weight * 0.8;
      }
    }
    return { link, score };
  }).filter(item => item.score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
    .map(item => item.link);
}

function candidateSummary(candidates) {
  return candidates.map((link, id) => ({
    id,
    name: String(link.name || '').slice(0, 110),
    url: String(link.url || '').slice(0, 220),
    category: String(link.category || '').slice(0, 70),
    section: String(link.section || '').slice(0, 90),
    tag: String(link.tag || '').slice(0, 70),
    note: String(link.note || '').slice(0, 140),
  }));
}

function parseModelResponse(content, candidates) {
  let raw = String(content || '').trim();
  const fenced = raw.match(/```(?:json)?\s*([\s\S]*?)```/i);
  if (fenced) raw = fenced[1].trim();
  let parsed = null;
  try { parsed = JSON.parse(raw); } catch (_) {
    const match = raw.match(/\{[\s\S]*\}/);
    if (match) {
      try { parsed = JSON.parse(match[0]); } catch (_) {}
    }
  }
  const picks = Array.isArray(parsed?.picks) ? parsed.picks : [];
  const seen = new Set();
  const results = picks.map(pick => {
    const id = Number(pick?.id);
    if (!Number.isInteger(id) || id < 0 || id >= candidates.length || seen.has(id)) return null;
    seen.add(id);
    return { link: candidates[id], reason: String(pick.reason || '').slice(0, 180) };
  }).filter(Boolean);
  return { answer: String(parsed?.answer || '').slice(0, 320), results };
}

async function advise(question, requestId) {
  const candidates = rankLinks(question);
  if (!candidates.length) {
    // Workers have no t(): fallback texts are sent as i18n keys (answerKey/reasonKey)
    // and translated by knowledge-advisor.js. Model output is passed as answer/reason.
    postMessage({ type: 'result', requestId, answerKey: 'kb_adv_no_match', results: [], fallback: true });
    return;
  }

  const catalog = candidateSummary(candidates);
  const system = `Jesteś doradcą stron w lokalnej Bazie Wiedzy. Polecaj WYŁĄCZNIE adresy z dostarczonej listy kandydatów. Nie wymyślaj stron ani URL-i. Wybierz maksymalnie 5 najtrafniejszych pozycji i krótko wyjaśnij dlaczego. Pisz po polsku z poprawnymi polskimi znakami diakrytycznymi (ą, ć, ę, ł, ń, ó, ś, ź, ż). Odpowiedz wyłącznie poprawnym JSON-em:
{"answer":"jedno krótkie zdanie po polsku","picks":[{"id":0,"reason":"krótki powód"}]}
Jeśli nic nie pasuje, zwróć {"answer":"Brak pewnej rekomendacji.","picks":[]}.
Kandydaci: ${JSON.stringify(catalog)}`;

  try {
    postMessage({ type: 'status', requestId, status: 'thinking' });
    if (!llmServer) throw new Error('Local AI server is not running');
    const response = await fetch(`${llmServer}/v1/chat/completions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages: [{ role: 'system', content: system }, { role: 'user', content: question }],
        temperature: 0.2,
        max_tokens: 300,
        stream: false,
      }),
    });
    if (!response.ok) throw new Error(`Model server: ${response.status}`);
    const data = await response.json();
    const parsed = parseModelResponse(data.choices?.[0]?.message?.content, candidates);
    const results = parsed.results.length ? parsed.results : candidates.slice(0, 5).map(link => ({ link, reasonKey: 'kb_adv_reason_match' }));
    if (parsed.answer) postMessage({ type: 'result', requestId, answer: parsed.answer, results, fallback: false });
    else postMessage({ type: 'result', requestId, answerKey: 'kb_adv_default_answer', results, fallback: false });
  } catch (_) {
    postMessage({ type: 'result', requestId, answerKey: 'kb_adv_model_unavailable', results: candidates.slice(0, 5).map(link => ({ link, reasonKey: 'kb_adv_reason_keyword' })), fallback: true });
  }
}

self.onmessage = event => {
  const message = event.data || {};
  if (message.serverUrl) llmServer = message.serverUrl;
  if (message.type === 'init') {
    links = Array.isArray(message.links) ? message.links : [];
    postMessage({ type: 'ready', total: links.length });
  }
  if (message.type === 'ask') advise(message.question || '', message.requestId || 0);
};
