// --- Page initialization ---
// === i18n init for subpage ===
// Counters and toggle buttons are rendered by JS (plural forms, current state),
// so they are refreshed here on load and after every language change.
function refreshDynamicLabels() {
    if (typeof updateCounts === 'function') updateCounts();
    if (typeof updateTweakCounts === 'function') updateTweakCounts();
    if (typeof updateDebloatCounts === 'function') updateDebloatCounts();
    if (typeof updateRemovedButton === 'function') updateRemovedButton();
    if (typeof updateActiveButton === 'function') updateActiveButton();
    if (typeof updateRemovedDebloatButton === 'function') updateRemovedDebloatButton();
}

document.addEventListener('DOMContentLoaded', () => {
    if (typeof applyTranslations === 'function') applyTranslations();
    refreshDynamicLabels();
    const langSelect = document.getElementById('lang-select');
    if (langSelect && typeof getCurrentLang === 'function') {
        langSelect.value = (typeof getLanguageChoice === "function" ? getLanguageChoice() : getCurrentLang());
        langSelect.addEventListener('change', (e) => {
            if (typeof setLanguage === 'function') setLanguage(e.target.value);
            if (typeof loadSystemInfo === 'function') loadSystemInfo();
            if (typeof renderProfiles === 'function') renderProfiles();
            if (typeof renderPrograms === 'function') renderPrograms();
            if (typeof renderTweaks === 'function') renderTweaks();
            if (typeof renderDebloat === 'function') renderDebloat();
            refreshDynamicLabels();
        });
    }
});
