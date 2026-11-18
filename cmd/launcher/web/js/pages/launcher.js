document.addEventListener('DOMContentLoaded', async () => {

// --- BAZA DANYCH #1: GŁÓWNA LISTA APLIKACJI I NARZĘDZI ---
    // Programy są teraz ładowane z API (programs.json) — patrz wywołanie
    // window.api.get_programs_data() tuż przed renderCategories()/renderItems().
    let items = [];

    // --- BAZA DANYCH #2: WSZYSTKIE STEROWNIKI ---
    // Katalog sterowników ładujemy z API (drivers.json) — patrz wywołanie
    // window.api.get_drivers_data() pod koniec tego handlera DOMContentLoaded.
    // driversData ma kształt: { version, types: { <typeKey>: { title, brands: { <brandKey>: { title, series: [...] } } } } }
    let driversData = { version: "1.0", types: {} };

    // --- BAZA DANYCH #3: PRZEWODNIK KONFIGURACJI ---
    const setupWorkflow = {
        get name() { return typeof t === 'function' ? t('workflow_title') : 'Pierwsza Konfiguracja Komputera'; },
        steps: [
            {
                get title() { return typeof t === 'function' ? t('workflow_step1_title') : 'Krok 1: Instalatory Zbiorcze (Ninite)'; },
                get description() { return typeof t === 'function' ? t('workflow_step1_desc') : 'Zainstaluj podstawowy zestaw programów za pomocą przygotowanych instalatorów Ninite.'; },
                subItems: [
                    { get name() { return typeof t === 'function' ? t('workflow_step1_sub1') : 'Pakiet bez narzędzi programistycznych'; }, folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/1. Ninite/programy bez programowania' },
                    { get name() { return typeof t === 'function' ? t('workflow_step1_sub2') : 'Pakiet z narzędziami programistycznymi'; }, folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/1. Ninite/programy z programowaniem' },
                    { get name() { return typeof t === 'function' ? t('workflow_step1_sub3') : 'Pakiet dla PC testowego'; }, folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/Instalacja automatyczna – PC testowy' }
                ]
            },
            {
                get title() { return typeof t === 'function' ? t('workflow_step2_title') : 'Krok 2: Debloat Systemu'; },
                get description() { return typeof t === 'function' ? t('workflow_step2_desc') : 'Uruchom skrypty PowerShell, aby usunąć zbędne komponenty i aplikacje z Windows.'; },
                folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/2. skrypt do debloata'
            },
            {
                get title() { return typeof t === 'function' ? t('workflow_step3_title') : 'Krok 3: Dodatkowe Programy'; },
                get description() { return typeof t === 'function' ? t('workflow_step3_desc') : 'Przejrzyj folder z najczęściej używanymi programami do ewentualnej, ręcznej instalacji.'; },
                folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/3. programy'
            },
            {
                get title() { return typeof t === 'function' ? t('workflow_step4_title') : 'Krok 4: Optymalizacja pod Gry'; },
                get description() { return typeof t === 'function' ? t('workflow_step4_desc') : 'Zastosuj modyfikacje rejestru, plany zasilania i inne tweaki w celu poprawy wydajności w grach.'; },
                folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/4. optymalizacja kompa pod gry'
            },
            {
                get title() { return typeof t === 'function' ? t('workflow_step5_title') : 'Krok 5: Personalizacja (Tapeta)'; },
                get description() { return typeof t === 'function' ? t('workflow_step5_desc') : 'Ustaw swoją ulubioną tapetę, aby zakończyć proces konfiguracji.'; },
                folderPath: '../Centrum Narzędzi i Oprogramowania/Programy do 1 konfiguracji komputera/5. tapeta'
            }
        ]
    };

    // --- CAŁA LOGIKA APLIKACJI (ZINTEGROWANA) ---
    const appContainer = document.getElementById('app-container');
    const template = document.getElementById('program-card-template');
    const categoryList = document.getElementById('category-list');
    const searchBox = document.getElementById('search-box');

    async function settleInitialProgramIcons() {
        await new Promise(resolve => requestAnimationFrame(resolve));
        const viewportLimit = window.innerHeight + 160;
        const visibleIcons = Array.from(appContainer.querySelectorAll('.program-icon'))
            .filter(icon => {
                const rect = icon.getBoundingClientRect();
                return rect.bottom >= 0 && rect.top <= viewportLimit;
            });

        const iconTasks = visibleIcons.map(icon => {
            if (icon.complete) {
                return typeof icon.decode === 'function' ? icon.decode().catch(() => {}) : Promise.resolve();
            }
            return new Promise(resolve => {
                icon.addEventListener('load', resolve, { once: true });
                icon.addEventListener('error', resolve, { once: true });
            });
        });

        await Promise.race([
            Promise.all(iconTasks),
            new Promise(resolve => setTimeout(resolve, 900))
        ]);
    }

    const mainView = document.getElementById('main-view');
    const detailView = document.getElementById('detail-view');
    const workflowView = document.getElementById('workflow-view');
    const detailContent = document.getElementById('detail-content');
    const workflowContent = document.getElementById('workflow-content');
    const backButtons = document.querySelectorAll('.back-btn');
    const helpButton = document.getElementById('help-btn');
    const helpModal = document.getElementById('help-modal');
    const closeModalButton = document.getElementById('close-modal-btn');

    let currentCategory = 'all';
    let aiModeActive = false;
    let aiResultItems = null;

    const tr = (key, fallback) => (typeof t === 'function' && t(key) !== key ? t(key) : fallback);

    const CARD_ICONS = {
        folder: '<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h4.1l2 2h8.9A1.5 1.5 0 0 1 21 8.5v9a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 17.5z" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/></svg>',
        copy: '<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><rect x="8" y="8" width="12" height="12" rx="2" fill="none" stroke="currentColor" stroke-width="1.7"/><path d="M16 8V5.5A1.5 1.5 0 0 0 14.5 4h-9A1.5 1.5 0 0 0 4 5.5v9A1.5 1.5 0 0 0 5.5 16H8" fill="none" stroke="currentColor" stroke-width="1.7"/></svg>',
        info: '<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><circle cx="12" cy="12" r="8.5" fill="none" stroke="currentColor" stroke-width="1.7"/><path d="M12 11v5.5M12 7.6v.4" stroke="currentColor" stroke-width="1.9" stroke-linecap="round"/></svg>',
        check: '<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M5 12.5l4.2 4.2L19 7" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>'
    };

    function programInitials(name) {
        const words = String(name || '')
            .replace(/\([^)]*\)/g, ' ')
            .split(/[\s\-_.\/]+/)
            .map(w => w.replace(/[^0-9A-Za-zÀ-ž]/g, ''))
            .filter(Boolean);
        if (words.length === 0) return '?';
        const letters = words.length > 1 ? words[0][0] + words[1][0] : words[0].slice(0, 2);
        return letters.toUpperCase();
    }

    function programHue(name) {
        let h = 0;
        for (const ch of String(name || '')) h = (h * 31 + ch.codePointAt(0)) >>> 0;
        return h % 360;
    }

    function makeInitialsAvatar(name) {
        const el = document.createElement('span');
        el.className = 'program-avatar';
        el.textContent = programInitials(name);
        el.style.setProperty('--avatar-h', String(programHue(name)));
        el.setAttribute('aria-hidden', 'true');
        return el;
    }

    // Icons that are missing, broken or unreadable are replaced by an
    // initials avatar. Results are cached per image URL.
    const iconVerdicts = new Map();

    // Heuristic tuned on a real 200-icon catalogue: an icon is replaced when
    // almost nothing in it stands out from the card background (dark artwork
    // on a dark card, transparent wordmarks), or when it is a wide, flat,
    // single-tone wordmark that is unreadable at 48 px.
    function iconIsUnreadable(img) {
        const w = img.naturalWidth, h = img.naturalHeight;
        if (!w || !h) return true;
        try {
            const size = 24;
            const scale = size / Math.max(w, h);
            const dw = Math.max(1, Math.round(w * scale)), dh = Math.max(1, Math.round(h * scale));
            const canvas = document.createElement('canvas');
            canvas.width = size; canvas.height = size;
            const ctx = canvas.getContext('2d', { willReadFrequently: true });
            ctx.drawImage(img, Math.floor((size - dw) / 2), Math.floor((size - dh) / 2), dw, dh);
            const data = ctx.getImageData(0, 0, size, size).data;
            const bg = document.body.classList.contains('light-theme') ? [255, 255, 255] : [22, 27, 36];
            let visible = 0, opaque = 0, sum = 0, sumSq = 0;
            for (let i = 0; i < data.length; i += 4) {
                if (data[i + 3] <= 40) continue;
                opaque++;
                const l = 0.2126 * data[i] + 0.7152 * data[i + 1] + 0.0722 * data[i + 2];
                sum += l; sumSq += l * l;
                const dr = data[i] - bg[0], dg = data[i + 1] - bg[1], db = data[i + 2] - bg[2];
                if (Math.sqrt(dr * dr + dg * dg + db * db) > 70) visible++;
            }
            if (visible / (size * size) < 0.03) return true;
            const mean = sum / Math.max(opaque, 1);
            const sd = Math.sqrt(Math.max(0, sumSq / Math.max(opaque, 1) - mean * mean));
            const aspect = Math.max(w / h, h / w);
            return aspect >= 1.5 && sd < 20;
        } catch (e) {
            return false;
        }
    }

    function useInitialsFallback(img, name) {
        if (!img || !img.isConnected) return;
        const wrap = img.closest('.program-icon-wrap') || img.parentElement;
        img.replaceWith(makeInitialsAvatar(name));
        if (wrap) wrap.classList.add('has-avatar');
    }

    function attachIconFallback(img, name, src) {
        if (!src) {
            useInitialsFallback(img, name);
            return;
        }
        const known = iconVerdicts.get(src);
        if (known === 'bad') {
            useInitialsFallback(img, name);
            return;
        }
        img.addEventListener('error', () => {
            iconVerdicts.set(src, 'bad');
            useInitialsFallback(img, name);
        }, { once: true });
        if (known !== 'ok') {
            img.addEventListener('load', () => {
                const verdict = iconIsUnreadable(img) ? 'bad' : 'ok';
                iconVerdicts.set(src, verdict);
                if (verdict === 'bad') useInitialsFallback(img, name);
            }, { once: true });
        }
        img.src = src;
    }
    window.coreDeckIconFallback = { attach: attachIconFallback, avatar: makeInitialsAvatar };

    function debounce(fn, delay = 160) {
        let timer = null;
        return (...args) => {
            clearTimeout(timer);
            timer = setTimeout(() => fn(...args), delay);
        };
    }

    function escapeHtmlValue(value) {
        return String(value ?? '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function detailLinkIcon(kind) {
        const icons = {
            guide: '↗',
            docs: '≡',
            download: '↓',
            official: '◎'
        };
        return icons[String(kind || '').toLowerCase()] || '↗';
    }

    const returnProgramStorageKey = 'launcher-return-program';

    function isExternalResource(url) {
        return /^(https?:|mailto:|lmstudio:)/i.test(String(url || ''));
    }

    function rememberReturnProgram(itemName) {
        if (!itemName) return;
        try {
            sessionStorage.setItem(returnProgramStorageKey, itemName);
        } catch (e) {
            console.warn('Nie udało się zapamiętać strony szczegółów:', e);
        }
    }

    function withReturnProgram(rawUrl, itemName) {
        try {
            const target = new URL(rawUrl, window.location.href);
            target.searchParams.set('returnProgram', itemName);
            return target.origin === window.location.origin
                ? `${target.pathname}${target.search}${target.hash}`
                : target.href;
        } catch (e) {
            console.warn('Nie udało się przygotować adresu strony programu:', e);
            return rawUrl;
        }
    }

    function openProgramResource(item, rawUrl, externalOverride) {
        const url = String(rawUrl || '').trim();
        if (!url) return;

        const external = typeof externalOverride === 'boolean'
            ? externalOverride
            : isExternalResource(url);

        if (external) {
            if (window.api && typeof window.api.open_url === 'function') {
                window.api.open_url(url);
            } else {
                window.open(url, '_blank', 'noopener');
            }
            return;
        }

        const itemName = item && item.name ? item.name : '';
        rememberReturnProgram(itemName);

        if (/^\/?user_pages\//i.test(url)) {
            const normalized = url.replace(/^\/+/, '');
            window.location.href = `/user-page-viewer.html?path=${encodeURIComponent(normalized)}&returnProgram=${encodeURIComponent(itemName)}`;
            return;
        }

        window.location.href = withReturnProgram(url, itemName);
    }

    function getProgramResourceLinks(item) {
        const links = Array.isArray(item.links)
            ? item.links.filter(link => link && link.url).map(link => ({ ...link }))
            : [];

        if (item.type !== 'link') return links;

        const directUrl = item.localPath || item.externalUrl;
        if (!directUrl || links.some(link => String(link.url) === String(directUrl))) return links;

        links.unshift({
            label: item.localPath ? 'Otwórz wbudowaną stronę' : 'Otwórz stronę projektu',
            url: directUrl,
            kind: item.localPath ? 'guide' : 'official'
        });
        return links;
    }

    function installTypeLabel(value) {
        const labels = {
            installer: 'Instalator',
            portable: 'Portable',
            script: 'Skrypt',
            tool: 'Narzędzie',
            addon: 'Dodatek',
            collection: 'Zbiór',
            other: 'Inne'
        };
        const key = String(value || '').toLowerCase();
        return labels[key] || value || 'Nie określono';
    }

    function getVariantList(item) {
        return Array.isArray(item.variants) ? item.variants.filter(Boolean) : [];
    }

    function getDefaultVariantIndex(item) {
        const variants = getVariantList(item);
        const configured = Number.isInteger(item.selectedVariant) ? item.selectedVariant : item.defaultVariant;
        const index = Number.isInteger(configured) ? configured : 0;
        return variants[index] ? index : 0;
    }

    function getProgramPath(item, variantIndex) {
        const variants = getVariantList(item);
        const variant = variants[variantIndex];
        return (variant && variant.folderPath) || item.folderPath || '';
    }

    function getProgramInstallType(item, variantIndex) {
        const variants = getVariantList(item);
        const variant = variants[variantIndex];
        return ((variant && variant.installType) || item.installType || '').toLowerCase();
    }

    function isPortableType(installType) {
        return installType === 'portable' || installType === 'tool';
    }

    function canRunProgram(item, variantIndex) {
        const installType = getProgramInstallType(item, variantIndex);
        return installType === 'installer' || installType === 'script' || isPortableType(installType);
    }

    function runLabelFor(installType) {
        if (isPortableType(installType)) return tr('run_program', 'Uruchom');
        if (installType === 'script') return tr('run_script', 'Uruchom skrypt');
        return tr('run_installer', 'Uruchom instalator');
    }

    function updateCardVariantPath(card) {
        const itemName = card.dataset.itemName || card.querySelector('.program-title')?.textContent;
        const item = items.find(i => i.name === itemName);
        if (!item) return;

        const variantSelect = card.querySelector('.program-variant-select');
        const variantIndex = variantSelect ? Number.parseInt(variantSelect.value, 10) : getDefaultVariantIndex(item);
        const path = getProgramPath(item, Number.isInteger(variantIndex) ? variantIndex : 0);

        card.querySelectorAll('.open-folder-btn, .copy-folder-btn').forEach(btn => {
            btn.dataset.path = path;
            btn.disabled = !path;
        });
        const detailsButton = card.querySelector('.details-btn');
        if (detailsButton) detailsButton.dataset.variantIndex = String(variantIndex);

        const runButton = card.querySelector('.run-program-btn');
        if (runButton) {
            const runnable = canRunProgram(item, variantIndex);
            const installType = getProgramInstallType(item, variantIndex);
            const label = runLabelFor(installType);
            runButton.hidden = !runnable;
            card.classList.toggle('is-runnable', runnable);
            runButton.dataset.variantIndex = String(variantIndex);
            runButton.dataset.defaultLabel = label;
            if (!runButton.disabled) runButton.textContent = `▶ ${label}`;
        }
    }

    function resetRunButton(button, defaultLabel, doneText) {
        if (!doneText) {
            button.disabled = false;
            button.textContent = `▶ ${defaultLabel}`;
            return;
        }
        button.textContent = doneText;
        setTimeout(() => {
            if (button.isConnected) {
                button.disabled = false;
                button.textContent = `▶ ${defaultLabel}`;
            }
        }, 2200);
    }

    // Translated text for a launch_program result (backend codes -> keys).
    function launchMessage(result) {
        if (result && result.code) {
            const key = `launch_${result.code}`;
            const text = tr(key, '');
            if (text) return fillText(text, result.params || {});
        }
        return (result && (result.error || result.message)) || tr('launch_failed', 'Nie udało się uruchomić programu.').replace('{error}', '');
    }

    // Translated text for an installer result (exit codes -> keys).
    function installerMessage(result) {
        const code = result && Number(result.exitCode);
        if (code) {
            const text = tr(`inst_code_${code}`, '');
            if (text) return text;
            if (result.outcome === 'error') return fillText(tr('inst_code_other', 'Instalator zakończył się błędem (kod {code}).'), { code });
        }
        return (result && result.message) || '';
    }

    function fillText(text, vars) {
        return String(text).replace(/\{([a-z_]+)\}/g, (m, name) => (name in vars ? String(vars[name]) : m));
    }

    // Portable programs and tools start in place, straight from the pendrive.
    async function launchPortable(button, item, variantIndex, defaultLabel) {
        if (!window.api || typeof window.api.launch_program !== 'function') {
            uiToast(tr('launch_unsupported', 'Uruchamianie programów portable wymaga nowszej wersji CoreDeck.'), { kind: 'error' });
            return;
        }
        button.disabled = true;
        button.textContent = `⏳ ${tr('run_starting', 'Uruchamianie…')}`;
        try {
            let result = await window.api.launch_program(item.name, variantIndex, '');
            if (result && Array.isArray(result.candidates) && result.candidates.length) {
                const choice = await uiChoose(
                    tr('launch_choose_hint', 'W folderze programu jest kilka plików .exe.'),
                    result.candidates.map(rel => {
                        const parts = rel.split('/');
                        return { label: parts[parts.length - 1], hint: parts.length > 1 ? rel : '', value: rel };
                    }),
                    {
                        title: `${tr('launch_choose_title', 'Który plik uruchomić?')} — ${item.name}`,
                        extra: { label: `📂 ${tr('launch_open_folder', 'Otwórz folder')}`, value: '__open_folder__' }
                    }
                );
                if (choice === '__open_folder__') {
                    const folderPath = getProgramPath(item, variantIndex);
                    addToHistory(item.name, folderPath);
                    openFolder(folderPath);
                    resetRunButton(button, defaultLabel);
                    return;
                }
                if (!choice) {
                    resetRunButton(button, defaultLabel);
                    return;
                }
                result = await window.api.launch_program(item.name, variantIndex, choice);
            }
            if (!result || result.error || !result.success) {
                throw new Error(launchMessage(result));
            }
            if (result.openedFolder) {
                uiToast(launchMessage(result), { kind: 'info', title: item.name });
                resetRunButton(button, defaultLabel);
                return;
            }
            uiToast(`${tr('launch_started', 'Uruchomiono')}: ${item.name}`, { kind: 'success', timeout: 2500 });
            resetRunButton(button, defaultLabel, `✓ ${tr('run_started_short', 'Uruchomiono')}`);
        } catch (error) {
            resetRunButton(button, defaultLabel);
            uiToast(error.message || String(error), { kind: 'error', title: item.name });
        }
    }

    // The click is the decision: installers and scripts start without asking.
    async function runProgramFromCard(button, card, item) {
        const variantIndex = Number.parseInt(button.dataset.variantIndex, 10);
        const selectedVariant = Number.isInteger(variantIndex) ? variantIndex : getDefaultVariantIndex(item);
        const installType = getProgramInstallType(item, selectedVariant);
        const defaultLabel = button.dataset.defaultLabel || 'Uruchom';

        if (isPortableType(installType)) {
            await launchPortable(button, item, selectedVariant, defaultLabel);
            return;
        }

        if (!window.api || typeof window.api.install_selected_programs !== 'function') {
            uiToast(tr('launch_unsupported', 'Funkcja uruchamiania nie jest dostępna w tej wersji programu.'), { kind: 'error' });
            return;
        }

        button.disabled = true;
        button.textContent = `⏳ ${tr('run_starting', 'Uruchamianie…')}`;

        try {
            const selected = Object.assign({}, item, { selectedVariant });
            const start = await window.api.install_selected_programs(JSON.stringify([selected]), '');
            if (start.error || !start.sessionId) {
                throw new Error(start.error || 'Nie udało się rozpocząć uruchamiania.');
            }

            let progress;
            do {
                await new Promise(resolve => setTimeout(resolve, 400));
                progress = await window.api.check_install_progress(start.sessionId);
                if (progress.error && !progress.done) throw new Error(progress.error);
            } while (!progress.done);

            const result = Array.isArray(progress.results) ? progress.results[0] : null;
            if (result && (result.cancelled || result.outcome === 'cancelled')) {
                resetRunButton(button, defaultLabel);
                uiToast(tr('install_cancelled', 'Instalacja anulowana'), { kind: 'neutral', title: item.name, timeout: 3000 });
                return;
            }
            if (!result || !result.success) {
                throw new Error(installerMessage(result) || progress.error || 'Uruchamianie zakończyło się błędem.');
            }

            if (result.outcome === 'ok_note') {
                uiToast(installerMessage(result), { kind: 'warn', title: item.name, timeout: 9000 });
            }
            resetRunButton(button, defaultLabel, `✓ ${tr('run_finished', 'Zakończono')}`);
        } catch (error) {
            resetRunButton(button, defaultLabel);
            uiToast(error.message || String(error), { kind: 'error', title: item.name });
        }
    }

    function renderItems() {
        appContainer.innerHTML = '';

        let filteredItems;
        if (aiModeActive) {
            if (aiResultItems === null) {
                appContainer.innerHTML = `<p style="padding:20px;color:var(--text-muted)">${typeof t === 'function' ? t('ai_placeholder') : 'Wpisz zapytanie i poczekaj na wyniki AI/LLM.'}</p>`;
                return;
            }
            filteredItems = aiResultItems
                .map(result => items.find(item => item.name === result.name) || result)
                .filter(Boolean);
        } else {
            const searchTerm = searchBox.value.toLowerCase();
            filteredItems = items.filter(item => {
                const inCategory = currentCategory === 'all' || item.category === currentCategory;
                const inSearch = (item.name || '').toLowerCase().includes(searchTerm) ||
                    (item.description || '').toLowerCase().includes(searchTerm) ||
                    (item.tags && item.tags.some(tag => tag.toLowerCase().includes(searchTerm)));
                return inCategory && inSearch;
            });
        }

        if (filteredItems.length === 0) { appContainer.innerHTML = `<p>${typeof t === 'function' ? t('no_results') : 'Brak pasujących programów.'}</p>`; return; }

        filteredItems.forEach(item => {
            const cardContent = template.content.cloneNode(true);
            const card = cardContent.querySelector('.program-card'); // Pobieramy element <a> z szablonu
            card.dataset.itemName = item.name;

            attachIconFallback(card.querySelector('.program-icon'), item.name, item.icon || '');
            card.querySelector('.program-title').textContent = item.name;
            card.querySelector('.program-title').title = item.name;
            card.querySelector('.program-description').textContent = item.description || '';
            if (item.description) card.querySelector('.program-description').title = item.description;

            const buttonsContainer = card.querySelector('.buttons');
            const variants = getVariantList(item);

            if (item.type === 'link') {
                card.dataset.url = item.localPath || item.externalUrl;
                card.dataset.external = item.externalUrl ? 'true' : 'false';
                const openLabel = item.externalUrl ? tr('open_page', 'Otwórz stronę') : tr('open_link', 'Otwórz');
                const detailsLabel = tr('details', 'Szczegóły');
                buttonsContainer.innerHTML = `
                    <button class="btn card-primary-btn open-link-btn">↗ ${escapeHtmlValue(openLabel)}</button>
                    <button class="card-icon-btn details-btn" title="${escapeHtmlValue(detailsLabel)}" aria-label="${escapeHtmlValue(detailsLabel)}">${CARD_ICONS.info}</button>
                `;
                card.querySelector('.details-btn').dataset.itemName = item.name;
            } else {
                if (variants.length > 0) {
                    const variantWrap = document.createElement('div');
                    variantWrap.className = 'program-variant';

                    const label = document.createElement('label');
                    label.textContent = typeof t === 'function' ? t('variant') : 'Wariant';

                    const select = document.createElement('select');
                    select.className = 'program-variant-select';
                    select.setAttribute('aria-label', `Wariant programu ${item.name}`);

                    variants.forEach((variant, index) => {
                        const defaultLabel = (typeof t === 'function' ? t('variant') : 'Wariant');
                        const option = new Option(variant.name || `${defaultLabel} ${index + 1}`, String(index));
                        select.add(option);
                    });
                    select.value = String(getDefaultVariantIndex(item));

                    variantWrap.append(label, select);
                    buttonsContainer.before(variantWrap);
                }

                // Jeśli to zwykły program, generujemy przyciski
                // One primary action (run, otherwise details); folder and path
                // are icon buttons revealed on hover or keyboard focus.
                const openLabel = escapeHtmlValue(tr('open_folder', 'Otwórz folder'));
                const copyLabel = escapeHtmlValue(tr('copy_path', 'Skopiuj ścieżkę'));
                const detailsLabel = escapeHtmlValue(tr('details', 'Szczegóły'));
                buttonsContainer.innerHTML = `
                    <button class="btn card-primary-btn run-program-btn" hidden></button>
                    <button class="btn card-secondary-btn details-btn" aria-label="${detailsLabel}" title="${detailsLabel}"><span class="details-icon">${CARD_ICONS.info}</span><span class="details-label">${detailsLabel}</span></button>
                    <span class="card-quick-actions">
                        <button class="card-icon-btn open-folder-btn" title="${openLabel}" aria-label="${openLabel}">${CARD_ICONS.folder}</button>
                        <button class="card-icon-btn copy-folder-btn" title="${copyLabel}" aria-label="${copyLabel}">${CARD_ICONS.copy}</button>
                    </span>
                `;
                card.querySelector('.details-btn').dataset.itemName = item.name;
                updateCardVariantPath(card);
            }

            appContainer.appendChild(cardContent);
        });
    }

    function renderCategories() {
        const allLabel = typeof t === 'function' ? t('all') : 'Wszystkie';
        const categories = [allLabel, ...new Set(items.map(item => item.category))];
        categoryList.innerHTML = '';
        categories.forEach(categoryName => {
            if (!categoryName) return; // Ignoruj puste kategorie
            const li = document.createElement('li');
            li.textContent = categoryName;
            li.dataset.category = categoryName === allLabel ? 'all' : categoryName;
            li.setAttribute('role', 'button');
            li.tabIndex = 0;
            if (li.dataset.category === 'all') { li.classList.add('active'); }
            categoryList.appendChild(li);
        });
    }

    function selectCategory(li) {
        if (!li || li.tagName !== 'LI') return;
        showMainView();
        document.querySelector('#category-list li.active')?.classList.remove('active');
        li.classList.add('active');
        currentCategory = li.dataset.category;
        renderItems();
    }

    function showDetailView(itemName, requestedVariantIndex) {
        const item = items.find(i => i.name === itemName);
        if (!item) return;

        const parsedVariantIndex = Number.parseInt(requestedVariantIndex, 10);
        const variantIndex = Number.isInteger(parsedVariantIndex) ? parsedVariantIndex : getDefaultVariantIndex(item);
        const folderPath = getProgramPath(item, variantIndex);
        const installType = getProgramInstallType(item, variantIndex);
        const variants = getVariantList(item);
        const selectedVariant = variants[variantIndex];
        const installerFile = (selectedVariant && selectedVariant.installerFile) || item.installerFile || '';
        const links = getProgramResourceLinks(item);
        const tags = Array.isArray(item.tags) ? item.tags.filter(Boolean).slice(0, 8) : [];
        const runnable = canRunProgram(item, variantIndex);
        const runLabel = runLabelFor(installType);

        const resourceLinks = links.map(link => {
            const url = String(link.url || '');
            const external = isExternalResource(url);
            return `
                <a class="detail-resource-link" href="${escapeHtmlValue(url)}" ${external ? 'data-external="true"' : ''}>
                    <span class="detail-resource-icon" aria-hidden="true">${detailLinkIcon(link.kind)}</span>
                    <span>
                        <strong>${escapeHtmlValue(link.label || url)}</strong>
                        <small>${external ? 'Otwiera się poza launcherem' : 'Wbudowana strona programu'}</small>
                    </span>
                    <span class="detail-resource-arrow" aria-hidden="true">→</span>
                </a>
            `;
        }).join('');

        const technicalRows = [
            ['Kategoria', item.category],
            ['Typ', installTypeLabel(installType)],
            installerFile ? ['Plik', installerFile] : null,
            selectedVariant && selectedVariant.name ? ['Wariant', selectedVariant.name] : null,
            variants.length > 1 ? ['Dostępne warianty', String(variants.length)] : null,
            item.runFromTemp || (selectedVariant && selectedVariant.runFromTemp) ? ['Uruchamianie', 'Kopia tymczasowa'] : null
        ].filter(Boolean).map(([label, value]) => `
            <div class="detail-fact">
                <dt>${escapeHtmlValue(label)}</dt>
                <dd>${escapeHtmlValue(value)}</dd>
            </div>
        `).join('');

        const tagMarkup = tags.map(tag => `<span class="detail-tag">${escapeHtmlValue(tag)}</span>`).join('');
        const description = escapeHtmlValue(item.description || 'Brak opisu.');
        const name = escapeHtmlValue(item.name);
        const category = escapeHtmlValue(item.category || 'Program');

        detailContent.dataset.itemName = item.name;
        detailContent.dataset.variantIndex = String(variantIndex);
        detailContent.innerHTML = `
            <div class="detail-hero">
                <span class="program-icon-wrap detail-icon-wrap"><img class="detail-program-icon" alt=""></span>
                <div class="detail-heading">
                    <div class="detail-eyebrow">${category}</div>
                    <h1>${name}</h1>
                    <p>${description}</p>
                    ${tagMarkup ? `<div class="detail-tags">${tagMarkup}</div>` : ''}
                </div>
            </div>

            <div class="detail-actions">
                ${runnable ? `<button class="btn detail-action detail-run-program" data-variant-index="${variantIndex}" data-default-label="${escapeHtmlValue(runLabel)}">▶ ${escapeHtmlValue(runLabel)}</button>` : ''}
                ${folderPath ? `<button class="btn detail-action detail-open-folder" data-path="${escapeHtmlValue(folderPath)}">📂 Otwórz folder</button>` : ''}
                ${folderPath ? `<button class="btn detail-action detail-copy-path" data-path="${escapeHtmlValue(folderPath)}">Kopiuj ścieżkę</button>` : ''}
            </div>

            <div class="detail-layout">
                <div class="detail-main-column">
                    ${item.details ? `
                        <section class="detail-section">
                            <h2>O programie</h2>
                            <div class="detail-richtext">${item.details}</div>
                        </section>
                    ` : ''}
                    ${resourceLinks ? `
                        <section class="detail-section">
                            <h2>Instrukcje i linki</h2>
                            <div class="detail-resources">${resourceLinks}</div>
                        </section>
                    ` : ''}
                </div>
                <aside class="detail-facts" aria-label="Dane techniczne programu">
                    <h2>Dane techniczne</h2>
                    <dl>${technicalRows}</dl>
                </aside>
            </div>
        `;
        attachIconFallback(detailContent.querySelector('.detail-program-icon'), item.name, item.icon || '');
        mainView.classList.add('hidden');
        detailView.classList.remove('hidden');
        const detailUrl = new URL(window.location.href);
        detailUrl.searchParams.set('details', item.name);
        history.replaceState(null, '', `${detailUrl.pathname}${detailUrl.search}${detailUrl.hash}`);
        detailView.scrollIntoView({ block: 'start' });
    }

    function showWorkflowView() {
        workflowContent.innerHTML = `<h1>${setupWorkflow.name}</h1>`;
        setupWorkflow.steps.forEach(step => {
            const stepElement = document.createElement('div');
            stepElement.className = 'workflow-step';

            let html = `<h3>${step.title}</h3><p>${step.description}</p>`;
            if (step.folderPath) {
                html += `<button class="btn copy-path-btn" data-path="${step.folderPath}">${typeof t === 'function' ? t('copy_path') : 'Skopiuj ścieżkę'}</button>`;
            }
            if (step.subItems) {
                step.subItems.forEach(sub => {
                    html += `<div class="workflow-subitem"><span>${sub.name}</span><button class="btn copy-path-btn" data-path="${sub.folderPath}">${typeof t === 'function' ? t('copy_path') : 'Skopiuj ścieżkę'}</button></div>`;
                });
            }
            stepElement.innerHTML = html;
            workflowContent.appendChild(stepElement);
        });
        mainView.classList.add('hidden');
        workflowView.classList.remove('hidden');
    }

    function showMainView() {
        detailView.classList.add('hidden');
        workflowView.classList.add('hidden');
        mainView.classList.remove('hidden');
        const mainUrl = new URL(window.location.href);
        mainUrl.searchParams.delete('details');
        history.replaceState(null, '', `${mainUrl.pathname}${mainUrl.search}${mainUrl.hash}`);
    }

    async function resolvePathForClipboard(path) {
        if (window.api && typeof window.api.resolve_path === 'function') {
            try {
                const resolved = await window.api.resolve_path(path);
                if (resolved) return resolved;
            } catch (e) {
                console.warn('Nie udało się rozwiązać ścieżki przez API:', e);
            }
        }

        // Pobierz literę dysku z parametru URL (przekazany przez Python)
        const urlParams = new URLSearchParams(window.location.search);
        let driveLetter = urlParams.get('drive') || '';
        // Go launcher przekazuje np. "L" w URL-u; Windows potrzebuje "L:".
        if (driveLetter && /^[A-Za-z]$/.test(driveLetter)) {
            driveLetter += ':';
        }

        // Rozwiąż ścieżkę relatywną do absolutnej
        const fullPathUrl = new URL(path, window.location.href);
        let windowsPath = decodeURI(fullPathUrl.pathname).substring(1).replace(/\//g, '\\');

        // Jeśli ścieżka nie zaczyna się od litery dysku, dodaj ją
        if (driveLetter && !windowsPath.match(/^[A-Za-z]:/)) {
            windowsPath = driveLetter + '\\' + windowsPath;
        }

        return windowsPath;
    }

    async function copyPathToClipboard(button) {
        const path = button.dataset.path;
        if (!path) return;

        const windowsPath = await resolvePathForClipboard(path);

        navigator.clipboard.writeText(windowsPath).then(() => {
            const copied = tr('copied', 'Skopiowano!');
            if (button.dataset.copying) return;
            button.dataset.copying = '1';
            const original = button.innerHTML;
            const originalTitle = button.getAttribute('title');
            if (button.classList.contains('card-icon-btn')) {
                button.innerHTML = CARD_ICONS.check;
                button.classList.add('is-done');
                button.setAttribute('title', copied);
            } else {
                button.textContent = copied;
            }
            setTimeout(() => {
                if (!button) return;
                button.innerHTML = original;
                button.classList.remove('is-done');
                if (originalTitle !== null) button.setAttribute('title', originalTitle);
                delete button.dataset.copying;
            }, 1500);
        });
    }

    // Eksportuj do użytku w ai-helper.js
    window.copyPathToClipboard = copyPathToClipboard;

    appContainer.addEventListener('change', (e) => {
        if (!e.target.classList.contains('program-variant-select')) return;
        const card = e.target.closest('.program-card');
        if (card) updateCardVariantPath(card);
    });

    appContainer.addEventListener('click', (e) => {
        // Kliknięcie w całą kartę (dla linków)
        const card = e.target.closest('.program-card');
        if (!card) return;

        const button = e.target.closest('button');

        // 1. Kliknięcie karty otwiera jej główny zasób.
        if (!button && card.dataset.url) {
            const item = items.find(i => i.name === card.dataset.itemName);
            openProgramResource(item, card.dataset.url, card.dataset.external === 'true');
            return;
        }

        // 2. Obsługa przycisków wewnątrz kart
        if (!button) return;

        const path = button.dataset.path;
        e.preventDefault();
        e.stopPropagation();

        if (button.classList.contains('copy-folder-btn')) {
            copyPathToClipboard(button);
        } else if (button.classList.contains('open-link-btn')) {
            const item = items.find(i => i.name === card.dataset.itemName);
            openProgramResource(item, card.dataset.url, card.dataset.external === 'true');
        } else if (button.classList.contains('run-program-btn')) {
            const item = items.find(i => i.name === card.dataset.itemName);
            if (item) runProgramFromCard(button, card, item);
        } else if (button.classList.contains('open-folder-btn')) {
            if (path) {
                const name = card.querySelector('.program-title').textContent;
                addToHistory(name, path);
                openFolder(path);
            }
        } else if (button.classList.contains('details-btn')) {
            showDetailView(button.dataset.itemName, button.dataset.variantIndex);
        }
    });

    detailContent.addEventListener('click', (e) => {
        const button = e.target.closest('button');
        const item = items.find(i => i.name === detailContent.dataset.itemName);

        if (button && button.classList.contains('detail-copy-path')) {
            copyPathToClipboard(button);
            return;
        }
        if (button && button.classList.contains('detail-open-folder')) {
            const path = button.dataset.path;
            if (path) {
                if (item) addToHistory(item.name, path);
                openFolder(path);
            }
            return;
        }
        if (button && button.classList.contains('detail-run-program') && item) {
            runProgramFromCard(button, null, item);
            return;
        }

        const link = e.target.closest('a[href]');
        if (!link) return;
        const href = link.getAttribute('href') || '';
        e.preventDefault();
        openProgramResource(item, href, link.dataset.external === 'true' || isExternalResource(href));
    });

    workflowContent.addEventListener('click', (e) => {
        const button = e.target.closest('button');
        if (button && button.classList.contains('copy-path-btn')) {
            copyPathToClipboard(button);
        }
    });

    categoryList.addEventListener('click', (e) => {
        selectCategory(e.target.closest('li'));
    });
    categoryList.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            selectCategory(e.target.closest('li'));
        }
    });

    backButtons.forEach(btn => btn.addEventListener('click', showMainView));
    searchBox.addEventListener('input', debounce(() => {
        if (aiModeActive) return;
        renderItems();
    }));
    helpButton.addEventListener('click', () => helpModal.classList.remove('hidden'));
    closeModalButton.addEventListener('click', () => helpModal.classList.add('hidden'));
    helpModal.addEventListener('click', (e) => { if (e.target === helpModal) { helpModal.classList.add('hidden'); } });
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !helpModal.classList.contains('hidden')) {
            helpModal.classList.add('hidden');
        }
    });

    // === ZARZĄDZANIE — przyciski w nagłówku prowadzą do dedykowanych stron ===
    // (HTML ma onclick na każdym przycisku — ten blok jest niepotrzebny.)

    const driverTypeSelect = document.getElementById('driver-type');
    const driverBrandSelect = document.getElementById('driver-brand');
    const driverSeriesSelect = document.getElementById('driver-series');
    const driverResultContainer = document.getElementById('driver-result-container');
    const driverCopyPathBtn = document.getElementById('driver-copy-path-btn');
    const driverOpenBtn = document.getElementById('driver-open-btn');
    const driverSearchBox = document.getElementById('driver-search-box');
    const driverSearchSection = document.getElementById('driver-search-section');
    function populateDriverTypeSelect() {
        // Wyczyść wszystko poza placeholderem, a potem dodaj typy z API.
        driverTypeSelect.innerHTML = `<option value="">${typeof t === 'function' ? t('driver_select_type') : 'Wybierz typ...'}</option>`;
        if (!driversData || !driversData.types) return;
        for (const typeKey in driversData.types) {
            const option = new Option(driversData.types[typeKey].title, typeKey);
            driverTypeSelect.add(option);
        }
    }
    populateDriverTypeSelect();
    function resetSelect(select, defaultText) {
        select.innerHTML = `<option value="">${defaultText}</option>`;
        select.disabled = true;
    }
    driverTypeSelect.addEventListener('change', () => {
        const selectedType = driverTypeSelect.value;
        const brandLabel = typeof t === 'function' ? t('driver_select_brand') : 'Wybierz producenta...';
        const seriesLabel = typeof t === 'function' ? t('driver_select_series') : 'Wybierz serię/model...';
        resetSelect(driverBrandSelect, brandLabel);
        resetSelect(driverSeriesSelect, seriesLabel);
        driverResultContainer.classList.add('hidden');
        if (selectedType && driversData.types[selectedType]) {
            const brands = driversData.types[selectedType].brands || {};
            for (const brandKey in brands) {
                const option = new Option(brands[brandKey].title, brandKey);
                driverBrandSelect.add(option);
            }
            driverBrandSelect.disabled = false;
        }
    });
    driverBrandSelect.addEventListener('change', () => {
        const selectedType = driverTypeSelect.value;
        const selectedBrand = driverBrandSelect.value;
        const seriesLabel = typeof t === 'function' ? t('driver_select_series') : 'Wybierz serię/model...';
        resetSelect(driverSeriesSelect, seriesLabel);
        driverResultContainer.classList.add('hidden');
        if (selectedBrand && driversData.types[selectedType] && driversData.types[selectedType].brands[selectedBrand]) {
            const seriesList = driversData.types[selectedType].brands[selectedBrand].series || [];
            seriesList.forEach(series => {
                const option = new Option(series.name, series.folderPath);
                driverSeriesSelect.add(option);
            });
            driverSeriesSelect.disabled = false;
        }
    });
    driverSeriesSelect.addEventListener('change', () => {
        const selectedPath = driverSeriesSelect.value;
        if (selectedPath) {
            driverCopyPathBtn.dataset.path = selectedPath;
            if (driverOpenBtn) driverOpenBtn.dataset.path = selectedPath;
            driverResultContainer.classList.remove('hidden');
        } else {
            driverResultContainer.classList.add('hidden');
        }
    });
    driverCopyPathBtn.addEventListener('click', (e) => {
        copyPathToClipboard(e.currentTarget);
    });
    if (driverOpenBtn) {
        driverOpenBtn.addEventListener('click', (e) => {
            const path = e.currentTarget.dataset.path;
            if (path) openFolder(path);
        });
    }
    if (driverSearchSection) {
        driverSearchSection.addEventListener('click', (e) => {
            const button = e.target.closest('button');
            if (!button) return;
            if (button.classList.contains('open-folder-btn')) {
                const path = button.dataset.path;
                if (path) openFolder(path);
            } else if (button.classList.contains('copy-folder-btn')) {
                copyPathToClipboard(button);
            }
        });
    }

    // --- SZYBKA WYSZUKIWARKA STEROWNIKÓW ---
    function searchDrivers(searchTerm) {
        if (!searchTerm || searchTerm.trim().length < 2) {
            driverSearchSection.classList.add('hidden');
            return;
        }

        const term = searchTerm.toLowerCase();
        const results = [];

        // Przeszukuj całą bazę sterowników
        for (const typeKey in driversData.types) {
            const type = driversData.types[typeKey];
            for (const brandKey in (type.brands || {})) {
                const brand = type.brands[brandKey];
                (brand.series || []).forEach(series => {
                    // Sprawdź czy termin pasuje do typu, marki lub serii
                    const matchesType = type.title.toLowerCase().includes(term);
                    const matchesBrand = brand.title.toLowerCase().includes(term);
                    const matchesSeries = series.name.toLowerCase().includes(term);

                    if (matchesType || matchesBrand || matchesSeries) {
                        results.push({
                            type: type.title,
                            brand: brand.title,
                            series: series.name,
                            path: series.folderPath
                        });
                    }
                });
            }
        }

        displayDriverSearchResults(results, searchTerm);
    }

    function displayDriverSearchResults(results, searchTerm) {
        const resultsContainer = driverSearchSection.querySelector('.driver-search-results');
        const titleElement = driverSearchSection.querySelector('.search-section-title');

        if (results.length === 0) {
            const noFound = typeof t === 'function' ? t('no_drivers_found') : 'Nie znaleziono sterowników dla';
            const tryOther = typeof t === 'function' ? t('try_other') : 'Spróbuj użyć innych słów kluczowych.';
            titleElement.textContent = `${noFound}: "${searchTerm}"`;
            resultsContainer.innerHTML = `<p style="color: var(--text-muted); padding: 10px;">${tryOther}</p>`;
            driverSearchSection.classList.remove('hidden');
            return;
        }

        const foundDrivers = typeof t === 'function' ? t('found_drivers') : 'Znaleziono sterowników';
        titleElement.textContent = `${foundDrivers}: ${results.length} (wyszukiwanie: "${searchTerm}")`;
        resultsContainer.innerHTML = '';

        results.forEach(result => {
            const card = document.createElement('div');
            card.className = 'driver-result-card';
            card.innerHTML = `
                <div class="driver-result-info">
                    <strong>${result.type}</strong> → ${result.brand} → ${result.series}
                </div>
                    <button class="btn open-folder-btn" data-path="${result.path}">📂 ${typeof t === 'function' ? t('open_folder') : 'Otwórz folder'}</button>
                    <button class="btn copy-path-btn copy-folder-btn" data-path="${result.path}">${typeof t === 'function' ? t('copy_path') : 'Skopiuj ścieżkę'}</button>
            `;
            resultsContainer.appendChild(card);
        });

        driverSearchSection.classList.remove('hidden');

        // Scroll do wyników
        driverSearchSection.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }

    // Event listener dla szybkiego wyszukiwania sterowników
    driverSearchBox.addEventListener('input', debounce((e) => {
        const searchTerm = e.target.value;

        // Wyczyść menu rozwijane gdy używamy szybkiego wyszukiwania
        if (searchTerm.trim().length > 0) {
            driverTypeSelect.value = '';
            const brandLabel = typeof t === 'function' ? t('driver_select_brand') : 'Wybierz producenta...';
            const seriesLabel = typeof t === 'function' ? t('driver_select_series') : 'Wybierz serię/model...';
            resetSelect(driverBrandSelect, brandLabel);
            resetSelect(driverSeriesSelect, seriesLabel);
            driverResultContainer.classList.add('hidden');
        }

        searchDrivers(searchTerm);
    }));

    // Ukryj wyniki wyszukiwania sterowników gdy zmienia się kategoria
    categoryList.addEventListener('click', () => {
        driverSearchSection.classList.add('hidden');
        driverSearchBox.value = '';
    });

    // Event listener dla przycisków "Skopiuj ścieżkę" w wynikach wyszukiwania sterowników
    driverSearchSection.addEventListener('click', (e) => {
        const button = e.target.closest('button');
        if (button && button.classList.contains('copy-path-btn')) {
            copyPathToClipboard(button);
        }
    });

    // Załaduj wszystkie dane RÓWNOLEGLE (programs + drivers + settings)
    const [dataResult, ddResult, sdResult] = await Promise.allSettled([
        window.api.get_programs_data(),
        window.api.get_drivers_data(),
        window.api.get_settings_data(),
    ]);

    if (dataResult.status === 'fulfilled' && dataResult.value && dataResult.value.programs) {
        items = dataResult.value.programs;
    } else {
        console.error('Nie udało się załadować programów:', dataResult.reason);
        items = [];
    }

    if (ddResult.status === 'fulfilled' && ddResult.value && ddResult.value.types) {
        driversData = ddResult.value;
    } else {
        console.error('Nie udało się załadować sterowników:', ddResult.reason);
    }
    populateDriverTypeSelect();

    // === SETTINGS ===
    let settingsData = { categories: [] };
    const settingsList = document.getElementById('settings-list');

    if (sdResult.status === 'fulfilled' && sdResult.value && sdResult.value.categories) {
        settingsData = sdResult.value;
    } else {
        console.error('Nie udało się załadować ustawień:', sdResult.reason);
    }

    function renderSettingsList() {
        if (!settingsList) return;
        settingsList.innerHTML = '';
        settingsData.categories.forEach((cat, idx) => {
            const li = document.createElement('li');
            li.textContent = cat.name;
            li.dataset.path = cat.folderPath;
            li.dataset.idx = idx;
            li.setAttribute('role', 'button');
            li.tabIndex = 0;
            settingsList.appendChild(li);
        });
    }

    const settingsResultContainer = document.getElementById('settings-result-container');
    const settingsOpenBtn = document.getElementById('settings-open-btn');
    const settingsCopyBtn = document.getElementById('settings-copy-btn');
    const settingsSelectedName = document.getElementById('settings-selected-name');
    const settingsSelectedPath = document.getElementById('settings-selected-path');
    let selectedSettingPath = null;

    function selectSetting(li) {
        if (!li || !li.dataset.path) return;
        const path = li.dataset.path;
        selectedSettingPath = path;

        // Stan aktywny na klikniętym elemencie (czyszczony z rodzeństwa)
        if (settingsList) {
            settingsList.querySelectorAll('li.is-active').forEach(other => {
                other.classList.remove('is-active');
                other.removeAttribute('aria-current');
            });
        }
        li.classList.add('is-active');
        li.setAttribute('aria-current', 'true');

        // Ścieżka dla OBU przycisków (open + copy)
        if (settingsOpenBtn) settingsOpenBtn.dataset.path = path;
        if (settingsCopyBtn) settingsCopyBtn.dataset.path = path;

        // Pokaż, co zostało wybrane
        if (settingsSelectedName) settingsSelectedName.textContent = li.textContent.trim();
        if (settingsSelectedPath) settingsSelectedPath.textContent = path;

        if (settingsResultContainer) {
            settingsResultContainer.classList.remove('hidden');
            settingsResultContainer.scrollIntoView({ block: 'nearest' });
        }
    }

    settingsList?.addEventListener('click', (e) => {
        selectSetting(e.target.closest('li'));
    });
    settingsList?.addEventListener('dblclick', (e) => {
        const li = e.target.closest('li');
        if (!li || !li.dataset.path) return;
        selectSetting(li);
        openFolder(li.dataset.path);
    });
    settingsList?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            selectSetting(e.target.closest('li'));
        }
    });

    settingsOpenBtn?.addEventListener('click', () => {
        if (selectedSettingPath) openFolder(selectedSettingPath);
    });
    settingsCopyBtn?.addEventListener('click', () => {
        if (selectedSettingPath) copyPathToClipboard(settingsCopyBtn);
    });

    renderSettingsList();

    renderCategories();
    renderItems();
    const requestedDetails = new URLSearchParams(window.location.search).get('details');
    if (requestedDetails) {
        showDetailView(requestedDetails);
    }
    coreDeckLoadingGate.setStatus('Loading icons…');
    settleInitialProgramIcons()
        .catch(() => {})
        .finally(() => {
            coreDeckLoadingGate.finish();
            // Report a finished update/rollback, then the daily update check.
            window.CoreUpdater?.startupCheck?.().catch(e => console.warn('update check failed', e));
        });

    // === LANGUAGE SELECTOR ===
    const langSelect = document.getElementById('lang-select');
    if (langSelect && typeof getCurrentLang === 'function') {
        // Set current value
        langSelect.value = (typeof getLanguageChoice === "function" ? getLanguageChoice() : getCurrentLang());

        langSelect.addEventListener('change', (e) => {
            const lang = e.target.value;
            if (typeof setLanguage === 'function') {
                setLanguage(lang);
                // Re-render categories with new language
                renderCategories();
                renderItems();
                if (typeof updateStats === 'function') updateStats();
                // Re-populate driver selects with new language
                populateDriverTypeSelect();
                const brandLabel = typeof t === 'function' ? t('driver_select_brand') : 'Wybierz producenta...';
                const seriesLabel = typeof t === 'function' ? t('driver_select_series') : 'Wybierz serię/model...';
                resetSelect(driverBrandSelect, brandLabel);
                resetSelect(driverSeriesSelect, seriesLabel);
            }
        });
    }

    // ============================================
    // NEW FEATURES
    // ============================================

    // === PLURALIZATION ===
    function plur(n, one, few, many) {
        if (n === 1) return `${n} ${one}`;
        if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) return `${n} ${few}`;
        return `${n} ${many}`;
    }

    function plurWord(n, one, few, many) {
        if (n === 1) return one;
        if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) return few;
        return many;
    }

    // === STATS ===
    function updateStats() {
        const totalPrograms = items.length;
        const categories = new Set(items.map(item => item.category).filter(c => c));
        document.getElementById('stats-total').textContent = totalPrograms;
        const programsLabel = document.getElementById('stats-programs-label');
        if (programsLabel) programsLabel.textContent = plurWord(totalPrograms, 'program', 'programy', 'programów');
        document.getElementById('stats-categories').textContent = categories.size;
    }
    updateStats();

    // === THEME TOGGLE ===
    const themeToggleBtn = document.getElementById('theme-toggle-btn');
    const themeIcon = themeToggleBtn.querySelector('.theme-icon');

    // Load saved theme
    const savedTheme = localStorage.getItem('launcher-theme');
    if (savedTheme === 'light') {
        document.body.classList.add('light-theme');
        themeIcon.textContent = '☀️';
    }

    themeToggleBtn.addEventListener('click', () => {
        document.body.classList.toggle('light-theme');
        const isLight = document.body.classList.contains('light-theme');
        themeIcon.textContent = isLight ? '☀️' : '🌙';
        localStorage.setItem('launcher-theme', isLight ? 'light' : 'dark');
    });

    // === HISTORY ===
    const historySection = document.getElementById('history-section');
    const historyList = document.getElementById('history-list');
    const MAX_HISTORY = 5;

    function loadHistory() {
        const history = JSON.parse(localStorage.getItem('launcher-history') || '[]');
        if (history.length > 0) {
            historySection.classList.remove('hidden');
            historyList.innerHTML = '';
            history.slice(0, MAX_HISTORY).forEach(item => {
                const li = document.createElement('li');
                li.textContent = item.name;
                li.dataset.path = item.path;
                li.title = item.path;
                historyList.appendChild(li);
            });
        } else {
            historySection.classList.add('hidden');
        }
    }
    loadHistory();

    function addToHistory(name, path) {
        let history = JSON.parse(localStorage.getItem('launcher-history') || '[]');
        // Remove duplicate if exists
        history = history.filter(h => h.path !== path);
        // Add to beginning
        history.unshift({ name, path });
        // Limit size
        history = history.slice(0, MAX_HISTORY);
        localStorage.setItem('launcher-history', JSON.stringify(history));
        loadHistory();
    }

    // Click on history item
    historyList.addEventListener('click', (e) => {
        if (e.target.tagName === 'LI') {
            const path = e.target.dataset.path;
            if (path) {
                openFolder(path);
            }
        }
    });

    // === OPEN FOLDER FUNCTION ===
    async function openFolder(path) {
        if (!path) return;

        console.log('[API] Natychmiastowe otwieranie:', path);

        const callApi = async () => {
            const api = window?.api;
            if (api && typeof api.open_folder === 'function') {
                try {
                    api.open_folder(path);
                    return true;
                } catch (e) {
                    return false;
                }
            }
            return false;
        };

        // Próba natychmiastowa
        let success = await callApi();

        // Jeśli nie od razu (np. API startup), ponów kilka razy po cichu w tle
        if (!success) {
            for (let i = 0; i < 3; i++) {
                await new Promise(r => setTimeout(r, 400));
                success = await callApi();
                if (success) break;
            }
        }

        if (!success) {
            navigator.clipboard.writeText(path);
        }
    }
    window.openFolder = openFolder;

    // Export programs and render hooks for AI helper
    window.launcherPrograms = items;
    window.setLauncherSearchMode = (mode) => {
        aiModeActive = mode === 'ai' || mode === 'llm';
        aiResultItems = null;
        renderItems();
    };
    window.renderLauncherAIResults = (programs) => {
        aiModeActive = true;
        aiResultItems = Array.isArray(programs) ? programs : [];
        renderItems();
    };

    // === APPLY i18n (AFTER all HTML is rendered) ===
    if (typeof applyTranslations === 'function') {
        applyTranslations();
    }
});
