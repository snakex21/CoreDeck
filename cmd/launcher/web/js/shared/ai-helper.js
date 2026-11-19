(function () {
    'use strict';

    const CONFIG = {
        maxResults: 10,
    };

    let worker = null;
    let isAIReady = false;
    let isLoading = false;
    let currentMode = 'normal';
    let loadedModel = null;
    let pendingQuery = '';
    let lastRequestId = 0;

    const elements = {
        searchBox: document.getElementById('search-box'),
        modeButtons: document.querySelectorAll('.mode-btn'),
        statusLoading: document.querySelector('.status-loading'),
        statusReady: document.querySelector('.status-ready'),
        statusError: document.querySelector('.status-error'),
        aiResultsSection: document.getElementById('ai-search-results'),
        aiResultsCount: document.getElementById('ai-results-count'),
        aiResultsContainer: document.querySelector('.ai-search-results-container'),
        clearAISearch: document.getElementById('clear-ai-search'),
    };

    // Tooltip on the warning icon so the user can see why AI search failed.
    function setErrorDetail(message) {
        if (!elements.statusError) return;
        if (message) {
            elements.statusError.title = message;
            elements.statusError.setAttribute('aria-label', message);
        } else {
            elements.statusError.removeAttribute('title');
            elements.statusError.removeAttribute('aria-label');
        }
    }

    // Worker messages carry translation keys; the worker has no t().
    function workerErrorText(msg) {
        if (msg.errorKey) return t(msg.errorKey, { error: msg.error || '' });
        return t('ai_err_search_failed', { error: msg.error || '' });
    }

    function showStatus(status) {
        if (status !== 'error') setErrorDetail('');
        elements.statusLoading?.classList.add('hidden');
        elements.statusReady?.classList.add('hidden');
        elements.statusError?.classList.add('hidden');
        if (status === 'loading') elements.statusLoading?.classList.remove('hidden');
        else if (status === 'ready') elements.statusReady?.classList.remove('hidden');
        else if (status === 'error') elements.statusError?.classList.remove('hidden');
    }

    function ensureWorker() {
        if (worker) return worker;
        worker = new Worker('./js/workers/ai-worker.js');

        worker.onmessage = (event) => {
            const msg = event.data || {};

            if (msg.type === 'status') {
                console.log('[LLM]', msg.status, msg.modelName || '');
                showStatus('loading');
            } else if (msg.type === 'ready') {
                isAIReady = true;
                isLoading = false;
                loadedModel = msg.model;
                showStatus('ready');
                console.log('[LLM] Ready:', msg.total, 'programs indexed');
                if (pendingQuery && pendingQuery.length >= 3) runSearch(pendingQuery);
            } else if (msg.type === 'result') {
                if (msg.requestId && msg.requestId !== lastRequestId) {
                    console.log('[LLM] Ignoring stale result', msg.requestId, 'current:', lastRequestId);
                    return;
                }
                showStatus('ready');
                displayResults(msg.results || [], msg.query || pendingQuery);
            } else if (msg.type === 'error') {
                if (msg.requestId && msg.requestId !== lastRequestId) return;
                console.error('[LLM ERROR]', msg.errorKey || '', msg.error);
                isAIReady = false;
                isLoading = false;
                showStatus('error');
                setErrorDetail(workerErrorText(msg));
            }
        };

        worker.onerror = (error) => {
            console.error('[LLM Worker fatal]', error);
            isAIReady = false;
            isLoading = false;
            showStatus('error');
            setErrorDetail(t('ai_err_worker_crashed'));
        };

        return worker;
    }

    async function initializeLLM() {
        if (isLoading) return;
        if (isAIReady) return;

        isLoading = true;
        isAIReady = false;
        loadedModel = null;
        showStatus('loading');

        try {
            if (!window.api?.ensure_llm_started) {
                throw new Error(t('ai_err_restart_after_update'));
            }
            const server = await window.api.ensure_llm_started();
            if (!server?.ready || !server.baseUrl) {
                throw new Error(tErr(server) || t('ai_err_start_failed'));
            }
            ensureWorker().postMessage({
                type: 'init',
                programs: window.launcherPrograms || [],
                serverUrl: server.baseUrl,
            });
        } catch (error) {
            console.error('[LLM START ERROR]', error);
            isLoading = false;
            isAIReady = false;
            showStatus('error');
            setErrorDetail(error?.message || String(error));
        }
    }

    function runSearch(query) {
        pendingQuery = query;
        if (!query || query.length < 3) return;

        if (!isAIReady) {
            initializeLLM();
            return;
        }

        showStatus('loading');
        const requestId = ++lastRequestId;
        ensureWorker().postMessage({
            type: 'search',
            query,
            maxResults: CONFIG.maxResults,
            requestId,
        });
    }

    function displayResults(results, query) {
        if (typeof window.renderLauncherAIResults === 'function') {
            window.renderLauncherAIResults((results || []).map(r => r.program).filter(Boolean));
        }

        if (!elements.aiResultsContainer) return;
        elements.aiResultsCount.textContent = results.length
            ? tn('count_ai_results_for', results.length, { query })
            : t('ai_no_results_for', { query });

        elements.aiResultsContainer.innerHTML = '';
        if (!results.length) {
            elements.aiResultsContainer.innerHTML = `<p style="padding:20px;text-align:center;color:var(--text-muted)">${t('no_ai_results')}</p>`;
        }

        results.forEach(({ program, similarity, reason }) => {
            const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
            const div = document.createElement('div');
            div.className = 'program-card';
            const path = esc(program.folderPath || program.localPath || '');
            const openLabel = esc(t('open_folder'));
            const copyLabel = esc(t('copy_path'));
            div.innerHTML = `
                <div class="program-card-head">
                    <span class="program-icon-wrap"><img alt="" class="program-icon" decoding="async"></span>
                    <div class="program-details">
                        <h3 class="program-title">${esc(program.name)}</h3>
                        <p class="program-description">${esc(program.description || '')}</p>
                    </div>
                </div>
                <div class="buttons">
                    <button class="btn card-primary-btn open-folder-btn" data-path="${path}">📂 ${openLabel}</button>
                    <span class="card-quick-actions">
                        <button class="card-icon-btn copy-path-btn" data-path="${path}" title="${copyLabel}" aria-label="${copyLabel}">⧉</button>
                    </span>
                </div>
            `;
            const img = div.querySelector('.program-icon');
            if (window.coreDeckIconFallback) {
                window.coreDeckIconFallback.attach(img, program.name, program.icon || '');
            } else {
                img.src = program.icon || 'icons/default.png';
            }
            elements.aiResultsContainer.appendChild(div);
        });

        elements.aiResultsSection?.classList.remove('hidden');
    }

    function attachResultButtonHandler() {
        if (!elements.aiResultsContainer || elements.aiResultsContainer.dataset.listenerAttached) return;
        elements.aiResultsContainer.addEventListener('click', (e) => {
            const button = e.target.closest('button');
            if (!button) return;
            const path = button.dataset.path;
            if (!path) return;
            e.preventDefault();
            e.stopPropagation();
            if (button.classList.contains('open-folder-btn')) {
                if (window.openFolder) window.openFolder(path);
            } else if (button.classList.contains('copy-path-btn')) {
                if (window.copyPathToClipboard) window.copyPathToClipboard(button);
            }
        });
        elements.aiResultsContainer.dataset.listenerAttached = 'true';
    }

    function init() {
        attachResultButtonHandler();

        elements.modeButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                elements.modeButtons.forEach(b => {
                    b.classList.remove('active');
                    b.setAttribute('aria-pressed', 'false');
                });
                btn.classList.add('active');
                btn.setAttribute('aria-pressed', 'true');
                currentMode = btn.dataset.mode;

                if (typeof window.setLauncherSearchMode === 'function') {
                    window.setLauncherSearchMode(currentMode);
                }

                if (currentMode === 'llm') {
                    initializeLLM();
                } else {
                    elements.aiResultsSection?.classList.add('hidden');
                }
            });
        });

        let timer;
        elements.searchBox?.addEventListener('input', (e) => {
            if (currentMode !== 'llm') return;
            clearTimeout(timer);
            const val = e.target.value.trim();
            if (val.length < 3) return;
            timer = setTimeout(() => runSearch(val), 400);
        });

        elements.clearAISearch?.addEventListener('click', () => {
            elements.searchBox.value = '';
            lastRequestId++;
            elements.aiResultsSection?.classList.add('hidden');
            if (typeof window.setLauncherSearchMode === 'function') {
                window.setLauncherSearchMode(currentMode);
            }
        });

        console.log('[LLM] Search mode ready');
    }

    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
    else init();
})();
