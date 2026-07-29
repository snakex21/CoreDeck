(() => {
  'use strict';

  const elements = {
    root: document.getElementById('kb-advisor'),
    toggle: document.getElementById('kb-advisor-toggle'),
    form: document.getElementById('kb-advisor-form'),
    input: document.getElementById('kb-advisor-input'),
    submit: document.getElementById('kb-advisor-submit'),
    results: document.getElementById('kb-advisor-results'),
  };
  if (!elements.root || !elements.form) return;

  let worker = null;
  let indexedData = null;
  let requestId = 0;

  function escapeHtml(value) {
    return String(value || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  function flattenKnowledge(data) {
    const groupNames = new Map((data?.groups || []).map(group => [group.id, group.name]));
    return (data?.categories || []).flatMap(category => (category.links || []).map(link => ({
      name: link.name || '', url: link.url || '', tag: link.tag || '', note: link.note || '',
      section: link.section || '', category: category.name || '', group: groupNames.get(category.groupId) || '',
    }))).filter(link => link.url);
  }

  function setStatus(message) {
    elements.results.innerHTML = `<div class="kb-advisor-status">${escapeHtml(message)}</div>`;
  }

  function ensureWorker(data) {
    if (!worker) {
      worker = new Worker('./js/workers/knowledge-advisor-worker.js');
      worker.onmessage = event => {
        const message = event.data || {};
        if (message.requestId && message.requestId !== requestId) return;
        if (message.type === 'status') setStatus('Doradca analizuje linki…');
        if (message.type === 'result') renderAnswer(message);
      };
      worker.onerror = () => {
        elements.submit.disabled = false;
        setStatus('Nie udało się uruchomić doradcy. Nadal możesz użyć zwykłej wyszukiwarki.');
      };
    }
    if (data && data !== indexedData) {
      indexedData = data;
      worker.postMessage({ type: 'init', links: flattenKnowledge(data) });
    }
    return worker;
  }

  function renderAnswer(message) {
    elements.submit.disabled = false;
    const picks = (message.results || []).map(item => {
      const link = item.link || {};
      const meta = [link.group, link.category, link.section].filter(Boolean).join(' · ');
      return `<article class="kb-advisor-pick">
        <div><div class="kb-advisor-pick-name">${escapeHtml(link.name)}</div>
        <div class="kb-advisor-pick-meta">${escapeHtml(meta)}</div>
        <div class="kb-advisor-pick-reason">${escapeHtml(item.reason || '')}</div></div>
        <button type="button" class="kb-advisor-open" data-url="${escapeHtml(link.url)}">Otwórz ↗</button>
      </article>`;
    }).join('');
    elements.results.innerHTML = `<div class="kb-advisor-message">${escapeHtml(message.answer || '')}</div>${picks || '<div class="kb-advisor-status">Brak pewnej rekomendacji — spróbuj doprecyzować pytanie.</div>'}`;
  }

  elements.toggle.addEventListener('click', () => {
    const isOpen = elements.root.classList.toggle('open');
    elements.toggle.setAttribute('aria-expanded', String(isOpen));
    elements.toggle.textContent = isOpen ? 'Zamknij doradcę' : 'Zapytaj doradcę';
    if (isOpen) elements.input.focus();
  });

  elements.form.addEventListener('submit', async event => {
    event.preventDefault();
    const question = elements.input.value.trim();
    if (question.length < 3) {
      setStatus('Napisz proszę trochę dokładniej, czego szukasz.');
      return;
    }
    const data = window.knowledgeData || (typeof knowledgeData !== 'undefined' ? knowledgeData : null);
    if (!data?.categories?.length) {
      setStatus('Baza Wiedzy nie jest jeszcze gotowa. Spróbuj ponownie za chwilę.');
      return;
    }
    elements.submit.disabled = true;
    setStatus('Uruchamiam lokalne AI…');
    requestId++;
    try {
      const server = await window.api?.ensure_llm_started?.();
      if (!server?.ready || !server.baseUrl) {
        throw new Error(server?.error || 'Lokalne AI nie jest dostępne.');
      }
      setStatus('Szukam pasujących stron w Twojej Bazie Wiedzy…');
      ensureWorker(data).postMessage({ type: 'ask', question, requestId, serverUrl: server.baseUrl });
    } catch (error) {
      elements.submit.disabled = false;
      setStatus(`Nie udało się uruchomić lokalnego AI: ${error.message || String(error)}`);
    }
  });

  elements.results.addEventListener('click', event => {
    const button = event.target.closest('.kb-advisor-open');
    if (!button?.dataset.url) return;
    if (typeof window.openKnowledgeUrl === 'function') window.openKnowledgeUrl(button.dataset.url);
  });

  window.addEventListener('knowledge-data-loaded', event => ensureWorker(event.detail));
})();
