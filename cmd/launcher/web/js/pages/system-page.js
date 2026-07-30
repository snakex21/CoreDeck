document.addEventListener('DOMContentLoaded', () => {
    const title = document.body.dataset.pageTitle || document.title;
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
        languageSelect.value = getCurrentLang();
        languageSelect.addEventListener('change', event => {
            if (typeof setLanguage === 'function') setLanguage(event.target.value);
            document.title = title;
        });
    }

    if (typeof window.initSystemTools === 'function') {
        window.initSystemTools();
    }

    const updateButton = document.getElementById('app-update-check');
    const openReleaseButton = document.getElementById('app-update-open');
    const updateStatus = document.getElementById('app-update-status');
    const currentVersion = document.getElementById('app-current-version');
    let releaseUrl = '';

    if (currentVersion && window.api?.get_app_version) {
        window.api.get_app_version().then(version => {
            currentVersion.textContent = version || '—';
        });
    }

    const setUpdateStatus = (message, state = '') => {
        if (!updateStatus) return;
        updateStatus.textContent = message;
        updateStatus.dataset.state = state;
    };

    updateButton?.addEventListener('click', async () => {
        if (!window.api?.check_for_updates) {
            setUpdateStatus('Sprawdzanie aktualizacji jest niedostępne w tej wersji.', 'error');
            return;
        }

        updateButton.disabled = true;
        openReleaseButton?.classList.add('hidden');
        setUpdateStatus('Łączenie z GitHub Releases…', 'checking');

        try {
            const result = await window.api.check_for_updates();
            if (currentVersion) currentVersion.textContent = result.currentVersion || '—';
            if (result.error) {
                setUpdateStatus(`Nie udało się sprawdzić: ${result.error}`, 'error');
                return;
            }
            if (result.updateAvailable) {
                releaseUrl = result.releaseUrl || '';
                setUpdateStatus(`Dostępna jest wersja ${result.latestVersion}.`, 'available');
                if (releaseUrl) openReleaseButton?.classList.remove('hidden');
                return;
            }
            setUpdateStatus(`Masz najnowszą wersję (${result.currentVersion}).`, 'current');
        } catch (error) {
            setUpdateStatus(`Nie udało się sprawdzić: ${String(error)}`, 'error');
        } finally {
            updateButton.disabled = false;
        }
    });

    openReleaseButton?.addEventListener('click', () => {
        if (releaseUrl) window.api?.open_url?.(releaseUrl);
    });
});
