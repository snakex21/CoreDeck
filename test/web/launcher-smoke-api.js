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

    // Any other binding resolves to an empty object instead of throwing.
    window.api = new Proxy(api, {
        get(target, prop) {
            if (prop in target) return target[prop];
            if (typeof prop !== 'string') return undefined;
            return async () => { calls.push({ m: prop }); return {}; };
        }
    });
})();
