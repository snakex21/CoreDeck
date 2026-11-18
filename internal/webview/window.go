package webview

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/webview/webview_go"

	"launcher/internal/api"
	"launcher/internal/system"
	"launcher/internal/windowstate"
)

type WindowConfig struct {
	Title     string
	URL       string
	Width     int
	Height    int
	MinWidth  int
	MinHeight int
	Maximized bool
	Debug     bool
	DataDir   string
	API       *api.API
	IconPath  string // absolute path to .ico file on disk
	StatePath string
}

func NewWindow(cfg WindowConfig) error {
	if cfg.DataDir != "" {
		os.MkdirAll(cfg.DataDir, 0755)
	}
	w := webview.New(cfg.Debug)
	defer w.Destroy()
	// The self-updater closes the app after starting its restart guard.
	cfg.API.SetQuitHook(func() {
		w.Dispatch(func() { w.Terminate() })
	})
	defer cfg.API.SetQuitHook(nil)
	w.SetTitle(cfg.Title)
	w.SetSize(cfg.Width, cfg.Height, webview.HintNone)
	w.SetSize(cfg.MinWidth, cfg.MinHeight, webview.HintMin)

	lastWindowState := windowstate.State{
		Width:     cfg.Width,
		Height:    cfg.Height,
		Maximized: cfg.Maximized,
	}
	w.Bind("__api_save_window_state", func() {
		if cfg.StatePath == "" {
			return
		}
		lastWindowState = captureWindowState(w, lastWindowState)
		if err := windowstate.Save(cfg.StatePath, lastWindowState); err != nil {
			slog.Warn("failed to save window state", "error", err)
		}
	})
	w.Bind("__api_open_folder", cfg.API.OpenFolder)
	w.Bind("__api_open_url", cfg.API.OpenURL)
	w.Bind("__api_get_available_models", cfg.API.GetAvailableModels)
	w.Bind("__api_ensure_llm_started", cfg.API.EnsureLLMStarted)
	w.Bind("__api_get_system_info", cfg.API.GetSystemInfo)
	w.Bind("__api_get_portable_paths", cfg.API.GetPortablePaths)
	w.Bind("__api_get_app_version", cfg.API.GetAppVersion)
	w.Bind("__api_get_update_settings", cfg.API.GetUpdateSettings)
	w.Bind("__api_set_update_settings", cfg.API.SetUpdateSettings)
	w.Bind("__api_check_for_updates", cfg.API.CheckForUpdates)
	w.Bind("__api_skip_update_version", cfg.API.SkipUpdateVersion)
	w.Bind("__api_start_update", cfg.API.StartUpdate)
	w.Bind("__api_update_progress", cfg.API.UpdateProgress)
	w.Bind("__api_update_ready", cfg.API.UpdateReady)
	w.Bind("__api_get_rollback_info", cfg.API.GetRollbackInfo)
	w.Bind("__api_rollback_update", cfg.API.RollbackUpdate)
	w.Bind("__api_get_programs_data", cfg.API.GetProgramsData)
	w.Bind("__api_get_programs_for_profile", cfg.API.GetProgramsForProfile)
	w.Bind("__api_install_selected_programs", cfg.API.InstallSelectedPrograms)
	w.Bind("__api_check_install_progress", cfg.API.CheckInstallProgress)
	w.Bind("__api_open_target_folder", cfg.API.OpenTargetFolder)
	w.Bind("__api_launch_program", cfg.API.LaunchProgram)
	w.Bind("__api_get_tweaks_data", cfg.API.GetTweaksData)
	w.Bind("__api_get_tweaks_for_profile", cfg.API.GetTweaksForProfile)
	w.Bind("__api_get_tweak_states", cfg.API.GetTweakStates)
	w.Bind("__api_start_tweak_state_scan", cfg.API.StartTweakStateScan)
	w.Bind("__api_check_tweak_state_scan", cfg.API.CheckTweakStateScan)
	w.Bind("__api_preview_tweaks", cfg.API.PreviewTweaks)
	w.Bind("__api_apply_tweaks", cfg.API.ApplyTweaks)
	w.Bind("__api_check_tweak_progress", cfg.API.CheckTweakProgress)
	w.Bind("__api_revert_tweak", cfg.API.RevertTweak)
	w.Bind("__api_get_custom_profiles", cfg.API.GetCustomProfiles)
	w.Bind("__api_save_custom_profile", cfg.API.SaveCustomProfile)
	w.Bind("__api_delete_custom_profile", cfg.API.DeleteCustomProfile)
	w.Bind("__api_add_program_to_profile", cfg.API.AddProgramToProfile)
	w.Bind("__api_remove_program_from_profile", cfg.API.RemoveProgramFromProfile)
	w.Bind("__api_add_program", cfg.API.AddProgram)
	w.Bind("__api_delete_program", cfg.API.DeleteProgram)
	w.Bind("__api_create_page", cfg.API.CreatePage)
	w.Bind("__api_delete_page", cfg.API.DeletePage)
	w.Bind("__api_get_drivers_data", cfg.API.GetDriversData)
	w.Bind("__api_save_drivers_data", cfg.API.SaveDriversData)
	w.Bind("__api_get_settings_data", cfg.API.GetSettingsData)
	w.Bind("__api_save_settings_data", cfg.API.SaveSettingsData)
	w.Bind("__api_get_knowledge_data", cfg.API.GetKnowledgeData)
	w.Bind("__api_save_knowledge_data", cfg.API.SaveKnowledgeData)
	w.Bind("__api_start_knowledge_link_check", cfg.API.StartKnowledgeLinkCheck)
	w.Bind("__api_get_knowledge_link_check", cfg.API.GetKnowledgeLinkCheck)
	w.Bind("__api_review_knowledge_link", cfg.API.ReviewKnowledgeLink)
	w.Bind("__api_get_knowledge_link_review_decision", cfg.API.GetKnowledgeLinkReviewDecision)
	w.Bind("__api_bulk_keep_knowledge_links", cfg.API.BulkKeepKnowledgeLinks)
	w.Bind("__api_open_knowledge_link_review", cfg.API.OpenKnowledgeLinkReview)
	w.Bind("__api_resolve_path", cfg.API.ResolvePath)
	w.Bind("__api_validate_export", cfg.API.ValidateExport)
	w.Bind("__api_export_pack", cfg.API.ExportPack)
	w.Bind("__api_import_pack", cfg.API.ImportPack)
	w.Bind("__api_preview_pack", cfg.API.PreviewPack)
	w.Bind("__api_list_export_history", cfg.API.ListExportHistory)
	w.Bind("__api_remove_export_history", cfg.API.RemoveExportHistory)
	w.Bind("__api_clear_export_history", cfg.API.ClearExportHistory)
	w.Bind("__api_pick_folder", cfg.API.PickFolder)
	w.Bind("__api_pick_open_file", cfg.API.PickOpenFile)
	w.Bind("__api_pick_save_file", cfg.API.PickSaveFile)
	w.Bind("__api_scan_installer_candidates", cfg.API.ScanInstallerCandidates)
	w.Bind("__api_cancel_operation", cfg.API.CancelOperation)

	w.Init(`
window.api = {
  save_window_state: function() { return __api_save_window_state().catch(function(){}) },
  open_folder: function(p) { return __api_open_folder(p).catch(function(e){console.error(e)}) },
  open_url: function(u) { return __api_open_url(u).catch(function(e){console.error(e)}) },
  get_available_models: function() { return __api_get_available_models().then(function(r){return JSON.parse(r)}).catch(function(e){return[]}) },
  ensure_llm_started: function() { return __api_ensure_llm_started().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_system_info: function() { return __api_get_system_info().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_portable_paths: function() { return __api_get_portable_paths().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_app_version: function() { return __api_get_app_version().catch(function(){return ''}) },
  get_update_settings: function() { return __api_get_update_settings().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  set_update_settings: function(j) { return __api_set_update_settings(typeof j === 'string' ? j : JSON.stringify(j||{})).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  check_for_updates: function(m) { return __api_check_for_updates(m === 'auto' ? 'auto' : 'manual').then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  skip_update_version: function(v) { return __api_skip_update_version(String(v||'')).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  start_update: function(v) { return __api_start_update(String(v||'')).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  update_progress: function(s) { return __api_update_progress(String(s||'')).then(function(r){return JSON.parse(r)}).catch(function(e){return{state:'error',error:String(e)}}) },
  update_ready: function() { return __api_update_ready().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_rollback_info: function() { return __api_get_rollback_info().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  rollback_update: function() { return __api_rollback_update().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_programs_data: function() { return __api_get_programs_data().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_programs_for_profile: function(p) { return __api_get_programs_for_profile(p).then(function(r){return JSON.parse(r)}).catch(function(e){return[]}) },
  install_selected_programs: function(p,t) { return __api_install_selected_programs(p,t).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  check_install_progress: function(s) { return __api_check_install_progress(s).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e),done:true}}) },
  open_target_folder: function(p) { return __api_open_target_folder(p).catch(function(e){console.error(e)}) },
  launch_program: function(n,v,f) { return __api_launch_program(String(n), Number.isInteger(v)?v:-1, f||'').then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_tweaks_data: function() { return __api_get_tweaks_data().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_tweaks_for_profile: function(p) { return __api_get_tweaks_for_profile(p).then(function(r){return JSON.parse(r)}).catch(function(e){return[]}) },
  get_tweak_states: function() { return __api_get_tweak_states().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  start_tweak_state_scan: function(f) { return __api_start_tweak_state_scan(!!f).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  check_tweak_state_scan: function(s) { return __api_check_tweak_state_scan(s).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e),done:true}}) },
  preview_tweaks: function(ids) { return __api_preview_tweaks(ids).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  apply_tweaks: function(t) { return __api_apply_tweaks(t).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  check_tweak_progress: function(s) { return __api_check_tweak_progress(s).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e),done:true}}) },
  revert_tweak: function(id) { return __api_revert_tweak(id).catch(function(e){return JSON.stringify({success:false,error:String(e)})}) },
  get_custom_profiles: function() { return __api_get_custom_profiles().then(function(r){return JSON.parse(r)}).catch(function(e){return[]}) },
  save_custom_profile: function(p) { return __api_save_custom_profile(p).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  delete_custom_profile: function(id) { return __api_delete_custom_profile(id).catch(function(e){return{error:String(e)}}) },
  add_program_to_profile: function(p,n) { return __api_add_program_to_profile(p,n).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  remove_program_from_profile: function(p,n) { return __api_remove_program_from_profile(p,n).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  add_program: function(p) { return __api_add_program(p).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  delete_program: function(n) { return __api_delete_program(n).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  create_page: function(r) { return __api_create_page(r).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  delete_page: function(r) { return __api_delete_page(r).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_drivers_data: function() { return __api_get_drivers_data().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  save_drivers_data: function(j) { return __api_save_drivers_data(j).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_settings_data: function() { return __api_get_settings_data().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  save_settings_data: function(j) { return __api_save_settings_data(j).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_knowledge_data: function() { return __api_get_knowledge_data().then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  save_knowledge_data: function(j) { return __api_save_knowledge_data(j).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  start_knowledge_link_check: function(c,l) { return __api_start_knowledge_link_check(c,l).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_knowledge_link_check: function(s) { return __api_get_knowledge_link_check(s).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e),done:true}}) },
  review_knowledge_link: function(r) { return __api_review_knowledge_link(r).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  get_knowledge_link_review_decision: function(c,u) { return __api_get_knowledge_link_review_decision(c,u).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  bulk_keep_knowledge_links: function(r) { return __api_bulk_keep_knowledge_links(r).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  open_knowledge_link_review: function(c,u) { return __api_open_knowledge_link_review(c,u).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) },
  resolve_path: function(p) { return Promise.resolve(__api_resolve_path(p)).catch(function(e){return p}) },
  validate_export: function(req) { return __api_validate_export(req).then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  export_pack: function(req) { return __api_export_pack(req).then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  import_pack: function(req) { return __api_import_pack(req).then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  preview_pack: function(p) { return __api_preview_pack(p).then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  list_export_history: function() { return __api_list_export_history().then(function(r){return JSON.parse(r)}).catch(function(e){return[]}) },
  remove_export_history: function(j) { return __api_remove_export_history(j).then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  clear_export_history: function() { return __api_clear_export_history().then(function(r){return JSON.parse(r)}).catch(function(e){return{success:false,error:String(e)}}) },
  pick_folder: function(t,i) { return __api_pick_folder(t,i).catch(function(e){return ''}) },
  pick_open_file: function(t,i,f) { return __api_pick_open_file(t,i,f).catch(function(e){return ''}) },
  pick_save_file: function(t,i,f,d) { return __api_pick_save_file(t,i,f,d).catch(function(e){return ''}) },
  scan_installer_candidates: function(fp) { return __api_scan_installer_candidates(fp).then(function(r){return JSON.parse(r)}).catch(function(e){return []}) },
  cancel_operation: function(s) { return __api_cancel_operation(s).then(function(r){return JSON.parse(r)}).catch(function(e){return{error:String(e)}}) }
};
if (/^\/(user_pages|knowledgebase)\//.test(window.location.pathname.toLowerCase())) {
  Object.keys(window).filter(function(k){return k.indexOf('__api_') === 0}).forEach(function(k){
    try { window[k] = undefined; delete window[k]; } catch (_) {}
  });
  window.api = Object.freeze({});
} else {
  var __coreDeckWindowStateTimer;
  window.addEventListener('resize', function() {
    clearTimeout(__coreDeckWindowStateTimer);
    __coreDeckWindowStateTimer = setTimeout(function() {
      window.api.save_window_state();
    }, 450);
  });
}
`)
	maximizeWindow(w, cfg.Maximized)
	w.Navigate(cfg.URL)

	setNativeWindowIconDelayed(w, cfg.IconPath)

	slog.Info("webview starting", "title", cfg.Title)
	w.Run()
	if cfg.StatePath != "" {
		if err := windowstate.Save(cfg.StatePath, lastWindowState); err != nil {
			slog.Warn("failed to save window state", "error", err)
		}
	}
	return nil
}

func BuildURL(port int, exeDir string) string {
	drive := system.DriveOf(exeDir)
	return fmt.Sprintf("http://localhost:%d/index.html?nocache=%d&drive=%s", port, time.Now().UnixMilli(), drive[:len(drive)-1])
}
