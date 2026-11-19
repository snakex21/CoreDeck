// ============================================
// INIT
// ============================================

document.addEventListener('DOMContentLoaded', async () => {
    initThemeToggle();
    initManageTabs();
    initProgramIconPicker();
    initPageIconPicker();
    await loadPrograms();
    await loadDrivers();
    await loadSettingsData();
    await loadKnowledgeData();
    initProgramsSearch();

    // Reakcja selecta serii na zmianę typu (dla modala edytora)
    const typeSel = document.getElementById('drv-series-type');
    const brandSel = document.getElementById('drv-series-brand');
    if (typeSel && brandSel) {
        typeSel.addEventListener('change', () => {
            const typeKey = typeSel.value;
            brandSel.innerHTML = `<option value="">${escapeHtml(t('mng_opt_choose'))}</option>`;
            if (!typeKey || !driversData.types[typeKey]) return;
            sortedBrandEntries(typeKey).forEach(([k, t]) => {
                brandSel.innerHTML += `<option value="${escapeHtml(k)}">${escapeHtml(t)}</option>`;
            });
        });
    }

    // ESC zamyka modale
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            // Zamknij overlay podglądu strony jeśli otwarty
            const previewOverlay = document.getElementById('page-preview-overlay');
            if (previewOverlay && previewOverlay.classList.contains('active')) {
                closePagePreview();
                return;
            }
            closeProgramEditor();
            closeDriverEditor();
            closeSettingEditor();
            closePageEditor();
        }
    });

    // Settings search
    const settingsSearchInput = document.getElementById('settings-search');
    if (settingsSearchInput) {
        settingsSearchInput.addEventListener('input', () => renderSettingsList());
    }

    // Knowledge search
    const knowledgeSearchInput = document.getElementById('knowledge-search-manage');
    if (knowledgeSearchInput) {
        knowledgeSearchInput.addEventListener('input', () => {
            activeKnowledgeGroupId = '';
            activeKnowledgeCategoryIndex = -1;
            renderKnowledgeList();
        });
    }

    // Pages search
    const pagesSearchInput = document.getElementById('pages-search-manage');
    if (pagesSearchInput) {
        pagesSearchInput.addEventListener('input', () => renderPagesList());
    }
});


// --- Page initialization ---
// === i18n init for subpage ===
document.addEventListener('DOMContentLoaded', () => {
    if (typeof applyTranslations === 'function') applyTranslations();
    const langSelect = document.getElementById('lang-select');
    if (langSelect && typeof getCurrentLang === 'function') {
        langSelect.value = (typeof getLanguageChoice === "function" ? getLanguageChoice() : getCurrentLang());
        langSelect.addEventListener('change', (e) => {
            if (typeof setLanguage === 'function') setLanguage(e.target.value);
        });
    }
});
