// Extra window.api stubs for launcher-smoke.html (load after smoke-api.js).
// Simulates the launcher's program actions without starting anything:
//  - portable entries: launch_program succeeds, or asks to choose when the
//    name contains "FMOD" (several .exe files in the folder);
//  - installers: the run ends with exit code 1602 (user cancelled) for
//    "ByClick", otherwise succeeds.
(function () {
    const api = window.api || (window.api = {});
    const calls = window.__smokeCalls || (window.__smokeCalls = []);
    const sessions = {};
    let counter = 0;

    api.launch_program = async (name, variant, file) => {
        calls.push({ m: 'launch_program', args: [name, variant, file] });
        if (/FMOD/i.test(name) && !file) {
            return { name, success: true, candidates: ['FMOD Studio.exe', 'fmodstudio.exe', 'fmodstudiocl.exe'] };
        }
        return { name, success: true, launched: 'L:\\Centrum\\' + name + '\\' + (file || name + '.exe') };
    };
    api.install_selected_programs = async (json) => {
        calls.push({ m: 'install_selected_programs', args: [json] });
        const id = 'inst-' + (++counter);
        sessions[id] = JSON.parse(json)[0];
        return { sessionId: id };
    };
    api.check_install_progress = async (id) => {
        const p = sessions[id];
        if (/ByClick/i.test(p.name)) {
            return { done: true, results: [{ name: p.name, success: false, cancelled: true, outcome: 'cancelled', exitCode: 1602, message: 'Instalacja anulowana', type: 'installer' }] };
        }
        return { done: true, results: [{ name: p.name, success: true, outcome: 'ok', message: 'Instalator zakończony', type: 'installer' }] };
    };

    // Updater stubs. ?update=1 offers version 9.9.9 with sample notes;
    // ?update=fail makes the download fail the checksum check.
    const updateMode = new URLSearchParams(location.search).get('update');
    let updateTick = 0;
    api.update_ready = async () => ({ justUpdated: '', rolledBack: null });
    api.check_for_updates = async (mode) => {
        calls.push({ m: 'check_for_updates', args: [mode] });
        if (!updateMode) return { currentVersion: '2.4.0', latestVersion: '2.4.0', updateAvailable: false };
        return {
            currentVersion: '2.4.0', latestVersion: '9.9.9', updateAvailable: true, prerelease: updateMode === 'test',
            publishedAt: '2026-11-18T12:00:00Z', channel: 'stable',
            notes: '## CoreDeck 9.9.9\n\n**PL**\n\n- **Aktualizacje w aplikacji**: kanał stabilny i testowy, suma SHA-256, kopia poprzedniej wersji.\n- **30 języków** interfejsu.\n\n**EN**\n\n- **In-app updates**: stable and test channels, SHA-256 check, previous version kept.\n- **30 interface languages**.'
        };
    };
    api.skip_update_version = async () => ({ success: true });
    api.start_update = async () => ({ sessionId: 'upd-1' });
    api.update_progress = async () => {
        updateTick += 1;
        if (updateMode === 'fail' && updateTick > 2) return { state: 'error', errorCode: 'bad_checksum', error: 'sha256 mismatch' };
        if (updateTick < 6) return { state: 'downloading', percent: updateTick * 18 };
        if (updateTick < 8) return { state: 'verifying' };
        return { state: 'restarting' };
    };

    // Any other binding resolves to an empty object instead of throwing.
    window.api = new Proxy(api, {
        get(target, prop) {
            if (prop in target) return target[prop];
            if (typeof prop !== 'string') return undefined;
            return async () => { calls.push({ m: prop }); return {}; };
        }
    });
})();
