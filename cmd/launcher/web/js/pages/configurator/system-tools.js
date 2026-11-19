(() => {
    let driversData = { version: '1.0', types: {} };
    let settingsData = { version: '1.0', categories: [] };
    let selectedDriverPath = '';
    let selectedSettingPath = '';
    let loadingPromise = null;
    // Status line renderers, re-run after a language change.
    const statusRenderers = new Map();

    function debounceSystemTool(fn, delay = 160) {
        let timer = null;
        return (...args) => {
            clearTimeout(timer);
            timer = setTimeout(() => fn(...args), delay);
        };
    }

    function systemEscapeHtml(value) {
        return String(value ?? '')
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    function countDriverSets() {
        let count = 0;
        Object.values(driversData.types || {}).forEach(type => {
            Object.values(type.brands || {}).forEach(brand => {
                count += Array.isArray(brand.series) ? brand.series.length : 0;
            });
        });
        return count;
    }

    async function resolveSystemPath(path) {
        if (window.api && typeof window.api.resolve_path === 'function') {
            try {
                return await window.api.resolve_path(path) || path;
            } catch (error) {
                console.warn('Nie udało się rozwiązać ścieżki:', error);
            }
        }
        return path;
    }

    async function openSystemPath(path) {
        if (!path || !window.api || typeof window.api.open_folder !== 'function') return;
        await window.api.open_folder(path);
    }

    async function copyText(text) {
        if (navigator.clipboard && window.isSecureContext) {
            try {
                await navigator.clipboard.writeText(text);
                return true;
            } catch (error) {
                console.warn('Clipboard API nie zadziałało:', error);
            }
        }

        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        let copied = false;
        try {
            copied = document.execCommand('copy');
        } finally {
            textarea.remove();
        }
        return copied;
    }

    async function copySystemPath(button, path) {
        if (!path) return;
        const resolved = await resolveSystemPath(path);
        const copied = await copyText(resolved);
        const original = button.textContent;
        button.textContent = copied ? t('copied') : t('err_copy_failed');
        setTimeout(() => {
            if (button) button.textContent = original;
        }, 1500);
    }

    // Empty first option of a select; data-i18n lets applyTranslations() relabel it.
    function placeholderOption(key) {
        const option = new Option(t(key), '');
        option.dataset.i18n = key;
        return option;
    }

    function resetSelect(select, key) {
        select.innerHTML = '';
        select.add(placeholderOption(key));
        select.disabled = true;
    }

    function populateDriverTypes() {
        const typeSelect = document.getElementById('config-driver-type');
        typeSelect.innerHTML = '';
        typeSelect.add(placeholderOption('driver_select_type'));
        Object.entries(driversData.types || {}).forEach(([key, type]) => {
            typeSelect.add(new Option(type.title || key, key));
        });
    }

    function collectDriverResults(searchTerm) {
        const normalized = String(searchTerm || '').trim().toLocaleLowerCase('pl');
        if (normalized.length < 2) return [];

        const results = [];
        Object.entries(driversData.types || {}).forEach(([typeKey, type]) => {
            Object.entries(type.brands || {}).forEach(([brandKey, brand]) => {
                (brand.series || []).forEach(series => {
                    const haystack = [
                        type.title,
                        brand.title,
                        series.name,
                        typeKey,
                        brandKey
                    ].join(' ').toLocaleLowerCase('pl');
                    if (!haystack.includes(normalized)) return;
                    results.push({
                        type: type.title || typeKey,
                        brand: brand.title || brandKey,
                        name: series.name,
                        path: series.folderPath
                    });
                });
            });
        });
        return results.slice(0, 100);
    }

    function renderDriverResults(results, searchTerm) {
        const section = document.getElementById('config-driver-search-section');
        const heading = document.getElementById('config-driver-results-heading');
        const list = section.querySelector('.config-driver-results-list');

        if (String(searchTerm || '').trim().length < 2) {
            section.classList.add('hidden');
            list.innerHTML = '';
            return;
        }

        section.classList.remove('hidden');
        if (results.length === 0) {
            heading.textContent = t('sys_drivers_no_results', { query: searchTerm });
            list.innerHTML = `<p class="system-tool-empty">${systemEscapeHtml(t('sys_drivers_no_results_hint'))}</p>`;
            return;
        }

        heading.textContent = tn('count_sys_drivers_found', results.length);
        list.innerHTML = results.map(result => `
            <article class="config-driver-result">
                <div>
                    <strong>${systemEscapeHtml(result.name)}</strong>
                    <small>${systemEscapeHtml(result.type)} · ${systemEscapeHtml(result.brand)}</small>
                </div>
                <div class="config-driver-result-actions">
                    <button class="btn-primary config-driver-result-open" data-path="${systemEscapeHtml(result.path)}">📂 ${systemEscapeHtml(t('open_link'))}</button>
                    <button class="btn-secondary config-driver-result-copy" data-path="${systemEscapeHtml(result.path)}">${systemEscapeHtml(t('sys_copy'))}</button>
                </div>
            </article>
        `).join('');
    }

    function initDriverControls() {
        const typeSelect = document.getElementById('config-driver-type');
        const brandSelect = document.getElementById('config-driver-brand');
        const seriesSelect = document.getElementById('config-driver-series');
        const searchInput = document.getElementById('config-driver-search');
        const searchSection = document.getElementById('config-driver-search-section');
        const selection = document.getElementById('config-driver-selection');
        const selectionName = document.getElementById('config-driver-selection-name');
        const selectionPath = document.getElementById('config-driver-selection-path');
        const openButton = document.getElementById('config-driver-open');
        const copyButton = document.getElementById('config-driver-copy');

        populateDriverTypes();

        typeSelect.addEventListener('change', () => {
            resetSelect(brandSelect, 'driver_select_brand');
            resetSelect(seriesSelect, 'driver_select_series');
            selection.classList.add('hidden');
            selectedDriverPath = '';

            const type = driversData.types[typeSelect.value];
            if (!type) return;
            Object.entries(type.brands || {}).forEach(([key, brand]) => {
                brandSelect.add(new Option(brand.title || key, key));
            });
            brandSelect.disabled = false;
        });

        brandSelect.addEventListener('change', () => {
            resetSelect(seriesSelect, 'driver_select_series');
            selection.classList.add('hidden');
            selectedDriverPath = '';

            const brand = driversData.types[typeSelect.value]?.brands?.[brandSelect.value];
            if (!brand) return;
            (brand.series || []).forEach(series => {
                const option = new Option(series.name, series.folderPath);
                option.dataset.name = series.name;
                seriesSelect.add(option);
            });
            seriesSelect.disabled = false;
        });

        seriesSelect.addEventListener('change', () => {
            const option = seriesSelect.selectedOptions[0];
            selectedDriverPath = seriesSelect.value;
            if (!selectedDriverPath) {
                selection.classList.add('hidden');
                return;
            }
            selectionName.textContent = option?.dataset.name || option?.textContent || '';
            selectionPath.textContent = selectedDriverPath;
            selection.classList.remove('hidden');
        });

        searchInput.addEventListener('input', debounceSystemTool(() => {
            renderDriverResults(collectDriverResults(searchInput.value), searchInput.value.trim());
        }));

        searchSection.addEventListener('click', event => {
            const open = event.target.closest('.config-driver-result-open');
            const copy = event.target.closest('.config-driver-result-copy');
            if (open) openSystemPath(open.dataset.path);
            if (copy) copySystemPath(copy, copy.dataset.path);
        });

        openButton.addEventListener('click', () => openSystemPath(selectedDriverPath));
        copyButton.addEventListener('click', () => copySystemPath(copyButton, selectedDriverPath));
    }

    function renderSettings(filter = '') {
        const list = document.getElementById('config-settings-list');
        const empty = document.getElementById('config-settings-empty');
        const normalized = filter.trim().toLocaleLowerCase('pl');
        const categories = (settingsData.categories || []).filter(item => {
            const haystack = `${item.name || ''} ${item.folderPath || ''}`.toLocaleLowerCase('pl');
            return !normalized || haystack.includes(normalized);
        });

        list.innerHTML = categories.map((item, index) => `
            <li
                role="button"
                tabindex="0"
                data-index="${index}"
                data-name="${systemEscapeHtml(item.name)}"
                data-path="${systemEscapeHtml(item.folderPath)}">
                ${systemEscapeHtml(item.name)}
            </li>
        `).join('');
        empty.classList.toggle('hidden', categories.length > 0);
    }

    function selectSetting(item) {
        if (!item || !item.dataset.path) return;
        const list = document.getElementById('config-settings-list');
        list.querySelectorAll('li.is-active').forEach(entry => entry.classList.remove('is-active'));
        item.classList.add('is-active');

        selectedSettingPath = item.dataset.path;
        document.getElementById('config-settings-name').textContent = item.dataset.name || item.textContent.trim();
        document.getElementById('config-settings-path').textContent = selectedSettingPath;
        document.getElementById('config-settings-placeholder').classList.add('hidden');
        document.getElementById('config-settings-selection').classList.remove('hidden');
    }

    function initSettingsControls() {
        const search = document.getElementById('config-settings-search');
        const list = document.getElementById('config-settings-list');
        const openButton = document.getElementById('config-settings-open');
        const copyButton = document.getElementById('config-settings-copy');

        renderSettings();
        search.addEventListener('input', debounceSystemTool(() => renderSettings(search.value)));
        list.addEventListener('click', event => selectSetting(event.target.closest('li')));
        list.addEventListener('dblclick', event => {
            const item = event.target.closest('li');
            if (!item) return;
            selectSetting(item);
            openSystemPath(item.dataset.path);
        });
        list.addEventListener('keydown', event => {
            if (event.key !== 'Enter' && event.key !== ' ') return;
            event.preventDefault();
            selectSetting(event.target.closest('li'));
        });
        openButton.addEventListener('click', () => openSystemPath(selectedSettingPath));
        copyButton.addEventListener('click', () => copySystemPath(copyButton, selectedSettingPath));
    }

    function setStatus(element, render) {
        if (!element) return;
        statusRenderers.set(element, render);
        element.removeAttribute('data-i18n'); // the initial "loading" label; JS owns the text from now on
        element.textContent = render();
    }

    // Re-renders the texts built in JS after the language picker changes.
    function refreshSystemToolTexts() {
        statusRenderers.forEach((render, element) => { element.textContent = render(); });
        const searchInput = document.getElementById('config-driver-search');
        if (searchInput && driversData.types) {
            renderDriverResults(collectDriverResults(searchInput.value), searchInput.value.trim());
        }
    }

    async function loadSystemTools() {
        const driversStatus = document.getElementById('drivers-data-status');
        const settingsStatus = document.getElementById('settings-data-status');
        setStatus(driversStatus, () => t('sys_loading_db'));
        setStatus(settingsStatus, () => t('sys_loading_db'));
        document.getElementById('lang-select')?.addEventListener('change', () => setTimeout(refreshSystemToolTexts, 0));
        const hasDriversPage = Boolean(document.getElementById('config-driver-type'));
        const hasSettingsPage = Boolean(document.getElementById('config-settings-list'));
        try {
            const [driversResult, settingsResult] = await Promise.allSettled([
                hasDriversPage ? window.api.get_drivers_data() : Promise.resolve(null),
                hasSettingsPage ? window.api.get_settings_data() : Promise.resolve(null)
            ]);

            if (hasDriversPage && driversResult.status === 'fulfilled' && driversResult.value?.types) {
                driversData = driversResult.value;
                initDriverControls();
                setStatus(driversStatus, () => tn('count_sys_drivers_sets', countDriverSets()));
            } else if (hasDriversPage) {
                setStatus(driversStatus, () => t('sys_db_load_failed'));
            }

            if (hasSettingsPage && settingsResult.status === 'fulfilled' && Array.isArray(settingsResult.value?.categories)) {
                settingsData = settingsResult.value;
                initSettingsControls();
                setStatus(settingsStatus, () => tn('count_sys_settings_sets', settingsData.categories.length));
            } else if (hasSettingsPage) {
                setStatus(settingsStatus, () => t('sys_settings_load_failed'));
            }
        } catch (error) {
            console.error('Nie udało się załadować narzędzi systemowych:', error);
            setStatus(driversStatus, () => t('sys_load_error'));
            setStatus(settingsStatus, () => t('sys_load_error'));
        }
    }

    window.initSystemTools = function initSystemTools() {
        if (!loadingPromise) loadingPromise = loadSystemTools();
        return loadingPromise;
    };
})();
