// --- Page initialization ---
// === i18n init for subpage ===
document.addEventListener('DOMContentLoaded', () => {
    if (typeof applyTranslations === 'function') applyTranslations();
    const langSelect = document.getElementById('lang-select');
    if (langSelect && typeof getCurrentLang === 'function') {
        langSelect.value = getCurrentLang();
        langSelect.addEventListener('change', (e) => {
            if (typeof setLanguage === 'function') setLanguage(e.target.value);
            if (typeof loadSystemInfo === 'function') loadSystemInfo();
            if (typeof renderProfiles === 'function') renderProfiles();
            if (typeof renderPrograms === 'function') renderPrograms();
            if (typeof renderTweaks === 'function') renderTweaks();
            if (typeof renderDebloat === 'function') renderDebloat();
            if (typeof updateCounts === 'function') updateCounts();
            if (typeof updateTweakCounts === 'function') updateTweakCounts();
            if (typeof updateDebloatCounts === 'function') updateDebloatCounts();
            if (typeof updateRemovedButton === 'function') updateRemovedButton();
        });
    }
});
