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

    function showStatus(status) {
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
                console.error('[LLM ERROR]', msg.error);
                isAIReady = false;
                isLoading = false;
                showStatus('error');
            }
        };

        worker.onerror = (error) => {
            console.error('[LLM Worker fatal]', error);
            isAIReady = false;
            isLoading = false;
            showStatus('error');
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
                throw new Error('Uruchom ponownie launcher po aktualizacji.');
            }
            const server = await window.api.ensure_llm_started();
            if (!server?.ready || !server.baseUrl) {
                throw new Error(server?.error || 'Nie udało się uruchomić lokalnego AI.');
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
            ? `${typeof t === 'function' ? t('results_for') : 'Znaleziono'} ${results.length} dla: "${query}"`
            : `${typeof t === 'function' ? t('no_ai_results') : 'Brak wyników dla:'} "${query}"`;

        elements.aiResultsContainer.innerHTML = '';
        if (!results.length) {
            elements.aiResultsContainer.innerHTML = `<p style="padding:20px;text-align:center;color:var(--text-muted)">${typeof t === 'function' ? t('no_ai_results') : 'Brak wyników.'}</p>`;
        }

        results.forEach(({ program, similarity, reason }) => {
            const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
            const div = document.createElement('div');
            div.className = 'program-card';
            const path = esc(program.folderPath || program.localPath || '');
            const openLabel = esc(typeof t === 'function' ? t('open_folder') : 'Otwórz folder');
            const copyLabel = esc(typeof t === 'function' ? t('copy_path') : 'Skopiuj ścieżkę');
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
