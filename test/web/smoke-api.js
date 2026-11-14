// Smoke-test stub for window.api.
// Serves the repository's local test data from /data/*.json over HTTP.
(function () {
    const BASE = '/data/';
    const cache = {};
    async function load(name) {
        if (!cache[name]) {
            const r = await fetch(BASE + name + '.json');
            if (!r.ok) throw new Error(name + ': HTTP ' + r.status);
            cache[name] = await r.json();
        }
        return cache[name];
    }
    window.__smokeCalls = [];
    function rec(m, args) { window.__smokeCalls.push({ m, args }); }

    window.api = {
        get_programs_data: async () => { rec('get_programs_data'); return await load('programs'); },
        get_drivers_data: async () => { rec('get_drivers_data'); return await load('drivers'); },
        get_settings_data: async () => { rec('get_settings_data'); return await load('settings'); },
        get_knowledge_data: async () => { rec('get_knowledge_data'); return await load('knowledge'); },
        // Go ListExportHistory marshals a slice -> bare array, not {entries:[]}
        list_export_history: async () => {
            rec('list_export_history');
            return window.__smokeHistory || [];
        },
        validate_export: async (j) => { rec('validate_export', j); return { estimatedSize: 12345678, warnings: [] }; },
        get_portable_paths: async () => { rec('get_portable_paths'); return { exeDir: 'J:\\_Launcher', drive: 'J:', root: 'J:\\', export: 'J:\\Export', import: 'J:\\Import', programs: 'J:\\Centrum Narzędzi i Oprogramowania', drivers: 'J:\\Centrum Narzędzi i Oprogramowania\\Drivers' }; },
        resolve_path: async (p) => { rec('resolve_path', p); return { path: 'J:\\' + p }; },
        open_url: async (u) => { rec('open_url', u); return { ok: true }; },
        open_folder: async (p) => { rec('open_folder', p); return { ok: true }; },
        pick_folder: async () => { rec('pick_folder'); return { path: '' }; },
        pick_open_file: async () => { rec('pick_open_file'); return { path: '' }; },
        pick_save_file: async () => { rec('pick_save_file'); return { path: '' }; },
        add_program: async () => ({ ok: true }),
        delete_program: async () => ({ ok: true }),
        create_page: async () => ({ ok: true }),
        delete_page: async () => ({ ok: true }),
        save_drivers_data: async () => ({ ok: true }),
        save_settings_data: async () => ({ ok: true }),
        save_knowledge_data: async () => ({ ok: true }),
        export_pack: async () => ({ ok: true }),
        import_pack: async () => ({ ok: true }),
        preview_pack: async () => ({ programs: [], drivers: [] }),
        clear_export_history: async () => ({ ok: true }),
        remove_export_history: async () => ({ ok: true }),
    };
})();
