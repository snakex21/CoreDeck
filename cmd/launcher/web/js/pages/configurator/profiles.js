// --- Custom Profiles ---
async function loadCustomProfiles() {
    try {
        customProfiles = await window.api.get_custom_profiles() || [];
        renderProfiles();
    } catch (e) {
        console.error('Failed to load custom profiles:', e);
    }
}

function renderProfiles() {
    const grid = document.getElementById('profiles-grid');
    const profileIcons = {
        common: '📦',
        laptop: '💻',
        ai_box: '🤖'
    };
    let html = '';

    // Built-in profiles
    for (const [key, profile] of Object.entries(profiles)) {
        // Check if there's a custom override
        const customOverride = customProfiles.find(cp => cp.id === key);
        const displayName = customOverride ? customOverride.name : profile.name;
        const displayDesc = customOverride ? customOverride.description : profile.description;
        const displayIcon = customOverride ? (customOverride.icon || profileIcons[key]) : profileIcons[key];
        const isOverridden = !!customOverride;

        // Count: use custom programs if overridden, otherwise built-in (including common)
        let count;
        if (customOverride && customOverride.programs) {
            count = customOverride.programs.length;
        } else {
            count = allPrograms.filter(p => p.profileRelevant && (p.profileRelevant.includes(key) || p.profileRelevant.includes('common'))).length;
        }

        html += `
            <div class="profile-card" onclick="applyProfile('${key}')" data-profile="${key}" ${isOverridden ? 'style="border-color:var(--accent-primary);"' : ''}>
                <div style="display:flex;justify-content:space-between;align-items:start;">
                    <div class="icon">${displayIcon || '⚙️'}</div>
                    <button onclick="event.stopPropagation();editBuiltinProfile('${key}')" style="background:var(--bg-active);border:1px solid var(--border-color);border-radius:var(--border-radius-sm);color:var(--text-secondary);cursor:pointer;padding:2px 6px;font-size:var(--fs-xs);" title="${t('title_edit')}">✏️</button>
                </div>
                <div class="name">${displayName}</div>
                <div class="desc">${displayDesc}</div>
                <div class="count">${tn('count_programs', count)}${isOverridden ? ' ' + t('cfg_profile_modified') : ''}</div>
            </div>
        `;
    }

    // Custom profiles
    for (const cp of customProfiles) {
        // Skip custom profiles that override built-in ones (already shown above)
        if (profiles[cp.id]) continue;

        const count = (cp.programs || []).length;
        html += `
            <div class="profile-card" onclick="applyCustomProfile('${cp.id}')" data-profile="custom_${cp.id}" style="border-style:dashed;">
                <div style="display:flex;justify-content:space-between;align-items:start;">
                    <div class="icon">${cp.icon || '📦'}</div>
                    <div style="display:flex;gap:4px;">
                        <button onclick="event.stopPropagation();editProfile('${cp.id}')" style="background:var(--bg-active);border:1px solid var(--border-color);border-radius:var(--border-radius-sm);color:var(--text-secondary);cursor:pointer;padding:2px 6px;font-size:var(--fs-xs);" title="${t('title_edit')}">✏️</button>
                        <button onclick="event.stopPropagation();deleteProfile('${cp.id}')" style="background:var(--bg-active);border:1px solid var(--border-color);border-radius:var(--border-radius-sm);color:var(--danger);cursor:pointer;padding:2px 6px;font-size:var(--fs-xs);" title="${t('title_delete')}">🗑️</button>
                    </div>
                </div>
                <div class="name">${cp.name}</div>
                <div class="desc">${cp.description || ''}</div>
                <div class="count">${tn('count_programs', count)}</div>
            </div>
        `;
    }

    grid.innerHTML = html;
}

async function applyProfile(profileName) {
    currentActiveProfile = profileName;
    currentActiveProfileIsCustom = false;
    document.querySelectorAll('.profile-card').forEach(c => c.classList.remove('selected'));
    document.querySelector(`[data-profile="${profileName}"]`).classList.add('selected');

    // Check for custom override
    const customOverride = customProfiles.find(cp => cp.id === profileName);

    // Select programs for profile
    allPrograms.forEach(p => p.selected = false);
    if (customOverride && customOverride.programs) {
        allPrograms.forEach(p => {
            if (customOverride.programs.includes(p.name)) {
                p.selected = true;
            }
        });
    } else {
        allPrograms.forEach(p => {
            if (p.profileRelevant && (p.profileRelevant.includes(profileName) || p.profileRelevant.includes('common'))) {
                p.selected = true;
                // Ustaw wariant per profil jeśli zdefiniowany
                if (p.profileVariants && typeof p.profileVariants[profileName] === 'number') {
                    p.selectedVariant = p.profileVariants[profileName];
                }
            }
        });
    }
    renderPrograms();

    // Select tweaks for profile
    allTweaks.forEach(t => t.selected = false);
    if (customOverride && customOverride.tweaks) {
        allTweaks.forEach(t => {
            if (customOverride.tweaks.includes(t.id)) {
                t.selected = true;
            }
        });
    } else {
        try {
            const profileTweaks = await window.api.get_tweaks_for_profile(profileName);
            const tweakIds = (profileTweaks || []).map(t => t.id);
            allTweaks.forEach(t => {
                t.selected = tweakIds.includes(t.id);
            });
        } catch (e) {
            console.error('Failed to load profile tweaks:', e);
        }
    }
    renderTweaks();

    document.querySelector('[data-tab="programs"]').click();
}

async function applyCustomProfile(profileId) {
    const cp = customProfiles.find(p => p.id === profileId);
    if (!cp) return;

    currentActiveProfile = profileId;
    currentActiveProfileIsCustom = true;
    document.querySelectorAll('.profile-card').forEach(c => c.classList.remove('selected'));
    document.querySelector(`[data-profile="custom_${profileId}"]`).classList.add('selected');

    // Select programs from custom profile
    allPrograms.forEach(p => p.selected = false);
    allPrograms.forEach(p => {
        if (cp.programs && cp.programs.includes(p.name)) {
            p.selected = true;
        }
    });
    renderPrograms();

    // Select tweaks from custom profile
    allTweaks.forEach(t => t.selected = false);
    if (cp.tweaks) {
        allTweaks.forEach(t => {
            if (cp.tweaks.includes(t.id)) {
                t.selected = true;
            }
        });
    }
    renderTweaks();

    document.querySelector('[data-tab="programs"]').click();
}

// --- Profile Editor ---
function createNewProfile() {
    editingProfileId = null;
    document.getElementById('profile-editor-title').textContent = t('cfg_new_profile');
    document.getElementById('profile-id').value = 'custom_' + Date.now();
    document.getElementById('profile-id').disabled = false;
    document.getElementById('profile-name').value = '';
    document.getElementById('profile-desc').value = '';
    document.getElementById('profile-icon').value = '📦';
    editorSelectedPrograms = [];
    editorSelectedTweaks = [];
    document.getElementById('profile-program-search').value = '';
    document.getElementById('profile-tweak-search').value = '';
    renderProfileProgramsList();
    renderProfileTweaksList();
    document.getElementById('profile-editor-overlay').classList.add('active');
}

function editProfile(profileId) {
    const cp = customProfiles.find(p => p.id === profileId);
    if (!cp) return;

    editingProfileId = profileId;
    document.getElementById('profile-editor-title').textContent = t('cfg_profile_editor');
    document.getElementById('profile-id').value = cp.id;
    document.getElementById('profile-id').disabled = true;
    document.getElementById('profile-name').value = cp.name;
    document.getElementById('profile-desc').value = cp.description || '';
    document.getElementById('profile-icon').value = cp.icon || '📦';
    editorSelectedPrograms = [...(cp.programs || [])];
    editorSelectedTweaks = [...(cp.tweaks || [])];
    document.getElementById('profile-program-search').value = '';
    document.getElementById('profile-tweak-search').value = '';
    renderProfileProgramsList();
    renderProfileTweaksList();
    document.getElementById('profile-editor-overlay').classList.add('active');
}

// Open the profile editor popup from the Tweaki tab.
// If a profile is currently selected, edit that profile's tweaks.
// If no profile is selected, alert the user.
function builtinProgramNamesForProfile(profileKey) {
    return allPrograms
        .filter(p => p.profileRelevant && (p.profileRelevant.includes(profileKey) || p.profileRelevant.includes('common')))
        .map(p => p.name);
}

function builtinTweakIdsForProfile(profileKey) {
    return allTweaks
        .filter(tw => tw.profiles && (tw.profiles.includes(profileKey) || tw.profiles.includes('common')))
        .map(tw => tw.id);
}

function tweakById(id) {
    return allTweaks.find(tw => tw.id === id);
}

function uniqueList(items) {
    return Array.from(new Set((items || []).filter(Boolean)));
}

function editCurrentProfileTweaks() {
    openProfileTweakEditor('tweaks');
}

function editCurrentProfileDebloat() {
    openProfileTweakEditor('debloat');
}

function openProfileTweakEditor(mode) {
    if (!currentActiveProfile) {
        uiToast(t('msg_select_profile_first'), { kind: 'warn' });
        return;
    }

    const profileKey = currentActiveProfile;
    const customOverride = customProfiles.find(cp => cp.id === profileKey);
    const builtinProfile = profiles[profileKey];
    const customProfile = customProfiles.find(cp => cp.id === profileKey);
    const profileName = customOverride?.name || builtinProfile?.name || customProfile?.name || profileKey;

    let selectedTweaks;
    if (customOverride && Array.isArray(customOverride.tweaks)) {
        selectedTweaks = customOverride.tweaks;
    } else if (builtinProfile) {
        selectedTweaks = builtinTweakIdsForProfile(profileKey);
    } else if (customProfile) {
        selectedTweaks = customProfile.tweaks || [];
    } else {
        selectedTweaks = [];
    }

    tweakEditorProfileId = profileKey;
    tweakEditorMode = mode === 'debloat' ? 'debloat' : 'tweaks';
    tweakEditorSelectedTweaks = uniqueList(selectedTweaks.filter(id => {
        const item = tweakById(id);
        return tweakEditorMode === 'debloat'
            ? item?.category === 'debloat'
            : item?.category !== 'debloat';
    }));

    const title = document.getElementById('tweak-profile-editor-title');
    if (title) title.textContent = `${t(tweakEditorMode === 'debloat' ? 'cfg_edit_profile_debloat' : 'cfg_edit_profile_tweaks')}: ${profileName}`;
    const hint = document.getElementById('tweak-profile-editor-hint');
    if (hint) hint.textContent = t(tweakEditorMode === 'debloat' ? 'cfg_debloat_profile_editor_hint' : 'cfg_tweak_profile_editor_hint');
    const search = document.getElementById('tweak-profile-search');
    if (search) search.value = '';
    renderTweakProfileEditorList();
    document.getElementById('tweak-profile-editor-overlay')?.classList.add('active');
}

function closeTweakProfileEditor() {
    document.getElementById('tweak-profile-editor-overlay')?.classList.remove('active');
    tweakEditorProfileId = null;
    tweakEditorSelectedTweaks = [];
    tweakEditorMode = 'tweaks';
}

function filterTweakProfileEditor() {
    renderTweakProfileEditorList();
}

function visibleTweaksForProfileEditor() {
    const searchInput = document.getElementById('tweak-profile-search');
    const filter = (searchInput?.value || '').toLowerCase();
    return allTweaks.filter(tw => {
        if (tweakEditorMode === 'debloat') {
            if (tw.category !== 'debloat') return false;
        } else if (tw.category === 'debloat') {
            return false;
        }
        const localizedName = tweakText(tw, 'name');
        const localizedDesc = tweakText(tw, 'description');
        return !filter ||
            localizedName.toLowerCase().includes(filter) ||
            localizedDesc.toLowerCase().includes(filter) ||
            (tw.category || '').toLowerCase().includes(filter) ||
            (tw.id || '').toLowerCase().includes(filter);
    });
}

function renderTweakProfileEditorList() {
    const container = document.getElementById('tweak-profile-list');
    if (!container) return;

    const visible = visibleTweaksForProfileEditor();
    const categoryNames = {
        power: t('cat_power'),
        privacy: t('cat_privacy'),
        visual: t('cat_visual'),
        services: t('cat_services'),
        edge: t('cat_edge'),
        gaming: t('cat_gaming'),
        security: t('cat_security'),
        debloat: t('cat_debloat')
    };

    let html = '';
    visible.forEach((tw) => {
        const checked = tweakEditorSelectedTweaks.includes(tw.id) ? 'checked' : '';
        const liveColor = tw.stateKnown && tw.active ? 'var(--success)' : (tw.stateKnown ? 'var(--warn)' : 'var(--text-muted)');
        const liveText = tw.stateKnown && tw.active ? t('tweak_active') : (tw.stateKnown ? t('tweak_inactive') : t('tweak_unchecked'));
        html += `
            <label style="display:flex;align-items:flex-start;gap:10px;padding:8px;border-radius:var(--border-radius-md);cursor:pointer;border:1px solid transparent;" onmouseover="this.style.background='var(--bg-hover)';this.style.borderColor='var(--border-color)'" onmouseout="this.style.background='transparent';this.style.borderColor='transparent'">
                <input type="checkbox" data-tweak-id="${tw.id}" ${checked} onchange="toggleTweakProfileEditor('${tw.id}', this.checked)" style="accent-color:var(--accent-primary);margin-top:3px;">
                <div style="min-width:0;flex:1;">
                    <div style="font-size:var(--fs-sm);color:var(--text-primary);font-weight:600;">${tweakText(tw, 'name')}</div>
                    <div style="font-size:var(--fs-xs);color:var(--text-muted);margin-top:2px;">${tweakText(tw, 'description')}</div>
                    <div style="display:flex;gap:6px;flex-wrap:wrap;margin-top:5px;">
                        <span style="font-size:var(--fs-xs);color:${liveColor};">● ${liveText}</span>
                        <span style="font-size:var(--fs-xs);color:var(--text-muted);">${categoryNames[tw.category] || tw.category || ''}</span>
                    </div>
                </div>
            </label>
        `;
    });
    container.innerHTML = html || `<p style="color:var(--text-muted);text-align:center;padding:12px;">${t('no_ai_results')}</p>`;

    const count = document.getElementById('tweak-profile-count');
    if (count) count.textContent = `${tweakEditorSelectedTweaks.length} ${t('cfg_selected')}`;
}

function toggleTweakProfileEditor(id, checked) {
    if (checked) {
        if (!tweakEditorSelectedTweaks.includes(id)) tweakEditorSelectedTweaks.push(id);
    } else {
        tweakEditorSelectedTweaks = tweakEditorSelectedTweaks.filter(twID => twID !== id);
    }
    renderTweakProfileEditorList();
}

function toggleAllVisibleTweakProfileEditor(checked) {
    const visibleIds = visibleTweaksForProfileEditor().map(tw => tw.id);
    if (checked) {
        tweakEditorSelectedTweaks = uniqueList([...tweakEditorSelectedTweaks, ...visibleIds]);
    } else {
        tweakEditorSelectedTweaks = tweakEditorSelectedTweaks.filter(id => !visibleIds.includes(id));
    }
    renderTweakProfileEditorList();
}

async function saveTweakProfileEditor() {
    const profileKey = tweakEditorProfileId;
    if (!profileKey) return;

    const builtinProfile = profiles[profileKey];
    const customProfile = customProfiles.find(cp => cp.id === profileKey);
    const baseTweaks = Array.isArray(customProfile?.tweaks)
        ? customProfile.tweaks
        : (builtinProfile ? builtinTweakIdsForProfile(profileKey) : []);
    const preservedOtherMode = baseTweaks.filter(id => {
        const item = tweakById(id);
        return tweakEditorMode === 'debloat'
            ? item?.category !== 'debloat'
            : item?.category === 'debloat';
    });
    const combinedTweaks = uniqueList([...preservedOtherMode, ...tweakEditorSelectedTweaks]);

    const profile = {
        id: profileKey,
        name: customProfile?.name || builtinProfile?.name || profileKey,
        description: customProfile?.description || builtinProfile?.description || '',
        icon: customProfile?.icon || '',
        programs: Array.isArray(customProfile?.programs)
            ? customProfile.programs
            : (builtinProfile ? builtinProgramNamesForProfile(profileKey) : []),
        tweaks: combinedTweaks
    };

    try {
        const result = await window.api.save_custom_profile(JSON.stringify(profile));
        if (result.error) {
            uiToast(t('err_prefix') + result.error, { kind: 'error' });
            return;
        }
        await loadCustomProfiles();
        closeTweakProfileEditor();
        if (profiles[profileKey]) await applyProfile(profileKey);
        else await applyCustomProfile(profileKey);
        renderProfiles();
    } catch (e) {
        uiToast(t('err_save_prefix') + String(e), { kind: 'error' });
    }
}

function editBuiltinProfile(profileKey) {
    const profile = profiles[profileKey];
    if (!profile) return;

    // Check if there's already a custom override
    const customOverride = customProfiles.find(cp => cp.id === profileKey);

    editingProfileId = profileKey;
    document.getElementById('profile-editor-title').textContent = t('cfg_profile_editor') + ': ' + profile.name;
    document.getElementById('profile-id').value = profileKey;
    document.getElementById('profile-id').disabled = true;
    document.getElementById('profile-name').value = customOverride ? customOverride.name : profile.name;
    document.getElementById('profile-desc').value = customOverride ? (customOverride.description || '') : profile.description;
    document.getElementById('profile-icon').value = customOverride ? (customOverride.icon || '📦') : '📦';

    // Get programs for this profile
    let selectedPrograms = [];
    if (customOverride && customOverride.programs) {
        selectedPrograms = customOverride.programs;
    } else {
        // Use built-in profile's programs
        selectedPrograms = allPrograms
            .filter(p => p.profileRelevant && (p.profileRelevant.includes(profileKey) || p.profileRelevant.includes('common')))
            .map(p => p.name);
    }
    editorSelectedPrograms = [...selectedPrograms];

    // Get tweaks for this profile
    let selectedTweaks = [];
    if (customOverride && customOverride.tweaks) {
        selectedTweaks = customOverride.tweaks;
    } else {
        // Use built-in profile's tweaks
        selectedTweaks = allTweaks
            .filter(t => t.profiles && (t.profiles.includes(profileKey) || t.profiles.includes('common')))
            .map(t => t.id);
    }
    editorSelectedTweaks = [...selectedTweaks];

    document.getElementById('profile-program-search').value = '';
    document.getElementById('profile-tweak-search').value = '';
    renderProfileProgramsList();
    renderProfileTweaksList();
    document.getElementById('profile-editor-overlay').classList.add('active');
}

function renderProfileProgramsList() {
    const container = document.getElementById('profile-programs-list');
    const searchInput = document.getElementById('profile-program-search');
    const filter = searchInput ? searchInput.value.toLowerCase() : '';

    let html = '';
    allPrograms.forEach((p, idx) => {
        const matchesFilter = !filter ||
            p.name.toLowerCase().includes(filter) ||
            (p.description || '').toLowerCase().includes(filter) ||
            (p.category || '').toLowerCase().includes(filter) ||
            (p.tags || []).some(t => t.toLowerCase().includes(filter));

        if (!matchesFilter) return;

        const checked = editorSelectedPrograms.includes(p.name) ? 'checked' : '';
        html += `
            <label style="display:flex;align-items:center;gap:8px;padding:4px 8px;cursor:pointer;border-radius:var(--border-radius-sm);" onmouseover="this.style.background='var(--bg-hover)'" onmouseout="this.style.background='transparent'">
                <input type="checkbox" data-program-name="${p.name}" ${checked} onchange="toggleEditorProgram('${p.name}', this.checked)" style="accent-color:var(--accent-primary);">
                <span style="font-size:var(--fs-sm);color:var(--text-primary);">${p.name}</span>
                <span style="font-size:var(--fs-xs);color:var(--text-muted);margin-left:auto;">${p.category}</span>
            </label>
        `;
    });
    container.innerHTML = html || `<p style="color:var(--text-muted);text-align:center;padding:12px;">${t('no_ai_results')}</p>`;
}

function toggleEditorProgram(name, checked) {
    if (checked) {
        if (!editorSelectedPrograms.includes(name)) {
            editorSelectedPrograms.push(name);
        }
    } else {
        editorSelectedPrograms = editorSelectedPrograms.filter(n => n !== name);
    }
}

function filterProfilePrograms() {
    renderProfileProgramsList();
}

// --- Profile Tweaks List ---
function renderProfileTweaksList() {
    const container = document.getElementById('profile-tweaks-list');
    const searchInput = document.getElementById('profile-tweak-search');
    const filter = searchInput ? searchInput.value.toLowerCase() : '';

    let html = '';
    allTweaks.forEach((tw) => {
        const localizedName = tweakText(tw, 'name');
        const localizedDesc = tweakText(tw, 'description');
        const matchesFilter = !filter ||
            localizedName.toLowerCase().includes(filter) ||
            localizedDesc.toLowerCase().includes(filter) ||
            (tw.category || '').toLowerCase().includes(filter);

        if (!matchesFilter) return;

        const checked = editorSelectedTweaks.includes(tw.id) ? 'checked' : '';
        html += `
            <label style="display:flex;align-items:center;gap:8px;padding:4px 8px;cursor:pointer;border-radius:var(--border-radius-sm);" onmouseover="this.style.background='var(--bg-hover)'" onmouseout="this.style.background='transparent'">
                <input type="checkbox" data-tweak-id="${tw.id}" ${checked} onchange="toggleEditorTweak('${tw.id}', this.checked)" style="accent-color:var(--accent-primary);">
                <span style="font-size:var(--fs-sm);color:var(--text-primary);">${localizedName}</span>
                <span style="font-size:var(--fs-xs);color:var(--text-muted);margin-left:auto;">${tw.category}</span>
            </label>
        `;
    });
    container.innerHTML = html || `<p style="color:var(--text-muted);text-align:center;padding:12px;">${t('no_ai_results')}</p>`;
}

function toggleEditorTweak(id, checked) {
    if (checked) {
        if (!editorSelectedTweaks.includes(id)) {
            editorSelectedTweaks.push(id);
        }
    } else {
        editorSelectedTweaks = editorSelectedTweaks.filter(t => t !== id);
    }
}

function filterProfileTweaks() {
    renderProfileTweaksList();
}

function closeProfileEditor() {
    document.getElementById('profile-editor-overlay').classList.remove('active');
    editingProfileId = null;
}

async function saveProfile() {
    const id = document.getElementById('profile-id').value.trim();
    const name = document.getElementById('profile-name').value.trim();
    const desc = document.getElementById('profile-desc').value.trim();
    const icon = document.getElementById('profile-icon').value.trim() || '📦';

    if (!id) { uiToast(t('val_profile_id'), { kind: 'warn' }); return; }
    if (!name) { uiToast(t('val_profile_name'), { kind: 'warn' }); return; }

    const profile = { id, name, description: desc, icon, programs: [...editorSelectedPrograms], tweaks: [...editorSelectedTweaks] };

    try {
        const result = await window.api.save_custom_profile(JSON.stringify(profile));
        if (result.error) {
            uiToast(t('err_prefix') + result.error, { kind: 'error' });
            return;
        }
        await loadCustomProfiles();
        closeProfileEditor();
    } catch (e) {
        uiToast(t('err_save_prefix') + String(e), { kind: 'error' });
    }
}

async function deleteProfile(profileId) {
    const cp = customProfiles.find(p => p.id === profileId);
    if (!cp) return;

    if (!(await uiConfirm(t('confirm_delete_profile') + ` "${cp.name}"?`, { danger: true }))) return;

    try {
        const result = await window.api.delete_custom_profile(profileId);
        if (result.error) {
            uiToast(t('err_prefix') + result.error, { kind: 'error' });
            return;
        }
        await loadCustomProfiles();
    } catch (e) {
        uiToast(t('err_delete_prefix') + String(e), { kind: 'error' });
    }
}
