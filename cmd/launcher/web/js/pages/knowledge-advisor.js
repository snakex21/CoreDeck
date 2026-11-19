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
        if (message.type === 'status') setStatus(t('kb_adv_status_thinking'));
        if (message.type === 'result') renderAnswer(message);
      };
      worker.onerror = () => {
        elements.submit.disabled = false;
        setStatus(t('kb_adv_worker_failed'));
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
        <div class="kb-advisor-pick-reason">${escapeHtml(item.reasonKey ? t(item.reasonKey) : (item.reason || ''))}</div></div>
        <button type="button" class="kb-advisor-open" data-url="${escapeHtml(link.url)}">${escapeHtml(t('open_link'))} ↗</button>
      </article>`;
    }).join('');
    // The worker sends answerKey for its own fallback texts; answer holds the model's own words.
    const answer = message.answerKey ? t(message.answerKey) : (message.answer || '');
    elements.results.innerHTML = `<div class="kb-advisor-message">${escapeHtml(answer)}</div>${picks || `<div class="kb-advisor-status">${escapeHtml(t('kb_adv_no_picks'))}</div>`}`;
  }

  elements.toggle.addEventListener('click', () => {
    const isOpen = elements.root.classList.toggle('open');
    elements.toggle.setAttribute('aria-expanded', String(isOpen));
    const toggleKey = isOpen ? 'kb_adv_toggle_close' : 'kb_adv_toggle_ask';
    elements.toggle.setAttribute('data-i18n', toggleKey);
    elements.toggle.textContent = t(toggleKey);
    if (isOpen) elements.input.focus();
  });

  elements.form.addEventListener('submit', async event => {
    event.preventDefault();
    const question = elements.input.value.trim();
    if (question.length < 3) {
      setStatus(t('kb_adv_too_short'));
      return;
    }
    const data = window.knowledgeData || (typeof knowledgeData !== 'undefined' ? knowledgeData : null);
    if (!data?.categories?.length) {
      setStatus(t('kb_adv_not_ready'));
      return;
    }
    elements.submit.disabled = true;
    setStatus(t('kb_adv_starting_ai'));
    requestId++;
    try {
      const server = await window.api?.ensure_llm_started?.();
      if (!server?.ready || !server.baseUrl) {
        throw new Error(tErr(server) || t('kb_adv_ai_unavailable'));
      }
      setStatus(t('kb_adv_searching'));
      ensureWorker(data).postMessage({ type: 'ask', question, requestId, serverUrl: server.baseUrl });
    } catch (error) {
      elements.submit.disabled = false;
      setStatus(t('kb_adv_ai_start_failed', { error: error.message || String(error) }));
    }
  });

  elements.results.addEventListener('click', event => {
    const button = event.target.closest('.kb-advisor-open');
    if (!button?.dataset.url) return;
    if (typeof window.openKnowledgeUrl === 'function') window.openKnowledgeUrl(button.dataset.url);
  });

  window.addEventListener('knowledge-data-loaded', event => ensureWorker(event.detail));
})();
