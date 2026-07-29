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
});
