document.addEventListener('DOMContentLoaded', () => {
    // applyTranslations() sets the generic "CoreDeck" title, so the page title
    // (data-page-title-key) is applied again after every translation pass.
    const titleKey = document.body.dataset.pageTitleKey;
    const applyPageTitle = () => {
        if (titleKey) document.title = t(titleKey);
    };
    applyTranslations();
    applyPageTitle();
    const themeButton = document.getElementById('theme-toggle-btn');
    const themeIcon = themeButton?.querySelector('.theme-icon');
    const languageSelect = document.getElementById('lang-select');

    if (localStorage.getItem('launcher-theme') === 'light') {
        document.body.classList.add('light-theme');
        if (themeIcon) themeIcon.textContent = '☀️';
    }

    themeButton?.addEventListener('click', () => {
        document.body.classList.toggle('light-theme');
        const light = document.body.classList.contains('light-theme');
        localStorage.setItem('launcher-theme', light ? 'light' : 'dark');
        if (themeIcon) themeIcon.textContent = light ? '☀️' : '🌙';
    });

    if (languageSelect && typeof getCurrentLang === 'function') {
        languageSelect.value = (typeof getLanguageChoice === "function" ? getLanguageChoice() : getCurrentLang());
        languageSelect.addEventListener('change', event => {
            if (typeof setLanguage === 'function') setLanguage(event.target.value);
            applyPageTitle();
        });
    }

    if (typeof window.initSystemTools === 'function') {
        window.initSystemTools();
    }

    // === Updates (Settings) ===
    const updateButton = document.getElementById('app-update-check');
    const rollbackButton = document.getElementById('app-update-rollback');
    const updateStatus = document.getElementById('app-update-status');
    const versionLine = document.getElementById('app-current-version-line');
    const channelSelect = document.getElementById('app-update-channel');
    const autoCheck = document.getElementById('app-update-auto');
    let installedVersion = '';
    let rollbackVersion = '';

    const setUpdateStatus = (message, state = '') => {
        if (!updateStatus) return;
        updateStatus.textContent = message;
        updateStatus.dataset.state = state;
    };
    const showVersion = () => {
        if (versionLine) versionLine.textContent = t('upd_installed_version', { version: installedVersion || '—' });
    };

    async function loadUpdateSettings() {
        if (window.api?.get_app_version) installedVersion = await window.api.get_app_version();
        showVersion();
        if (window.api?.get_update_settings) {
            const s = await window.api.get_update_settings();
            if (s && !s.error) {
                if (channelSelect) channelSelect.value = s.channel === 'test' ? 'test' : 'stable';
                if (autoCheck) autoCheck.checked = s.autoCheck !== false;
            }
        }
        if (window.api?.get_rollback_info && rollbackButton) {
            const info = await window.api.get_rollback_info();
            rollbackVersion = (info && info.available && info.version) || '';
            rollbackButton.classList.toggle('hidden', !rollbackVersion);
        }
    }

    async function saveUpdateSettings() {
        if (!window.api?.set_update_settings) return;
        const result = await window.api.set_update_settings(JSON.stringify({
            channel: channelSelect?.value === 'test' ? 'test' : 'stable',
            autoCheck: !!autoCheck?.checked
        }));
        if (result && result.error) uiToast(tErr(result), { kind: 'error' });
    }

    channelSelect?.addEventListener('change', saveUpdateSettings);
    autoCheck?.addEventListener('change', saveUpdateSettings);

    updateButton?.addEventListener('click', async () => {
        if (!window.CoreUpdater) return;
        updateButton.disabled = true;
        setUpdateStatus(t('upd_checking'), 'checking');
        try {
            const result = await window.CoreUpdater.checkNow();
            if (!result || result.error) {
                setUpdateStatus(t('upd_check_error', { error: (result && result.error) || '?' }), 'error');
            } else if (result.updateAvailable) {
                setUpdateStatus(t('upd_message', { version: result.latestVersion, current: result.currentVersion }), 'available');
            } else {
                setUpdateStatus(t('upd_up_to_date', { version: result.currentVersion || installedVersion }), 'current');
            }
        } catch (error) {
            setUpdateStatus(t('upd_check_error', { error: String(error) }), 'error');
        } finally {
            updateButton.disabled = false;
        }
    });

    rollbackButton?.addEventListener('click', async () => {
        if (!rollbackVersion || !window.api?.rollback_update) {
            uiToast(t('upd_rollback_none'), { kind: 'neutral' });
            return;
        }
        const ok = await uiConfirm(t('upd_rollback_confirm', { version: rollbackVersion }), { confirmLabel: t('upd_rollback') });
        if (!ok) return;
        const result = await window.api.rollback_update();
        if (result && result.error) uiToast(t('upd_failed', { error: result.error }), { kind: 'error' });
        else uiToast(t('upd_restarting'), { kind: 'success' });
    });

    document.getElementById('lang-select')?.addEventListener('change', () => setTimeout(showVersion, 0));
    loadUpdateSettings().catch(e => console.warn('update settings', e));
});
