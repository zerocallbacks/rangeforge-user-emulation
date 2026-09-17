package manager

import (
	"net/http"
)

// DashboardHTML contains the complete RangeForge open-source dashboard.
const DashboardHTML = `<!DOCTYPE html>
<html lang="en" data-theme="ocean-sapphire">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>RANGEFORGE // COMMAND CENTER</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <style>
    /* ==========================================================================
       RANGEFORGE // ENTERPRISE CYBER RANGE OPERATIONS CONSOLE
       Theme 1: Ocean Sapphire (Cobalt Deep Blue Console)
       Theme 2: Carbon Black (Stealth Dark Operations with Clean Slate Borders)
       ========================================================================== */

    /* THEME 1: OCEAN SAPPHIRE / COBALT OPERATIONS CONSOLE (RICH DEEP BLUE) */
    :root, [data-theme="ocean-sapphire"], [data-theme="cobalt-ops"], [data-theme="blue-ops"] {
      --bg-void: #061122;
      --bg-base: #0a1832;
      --bg-sidebar: #08152c;
      --bg-header: #0d1f40;
      --bg-card: #10254c;
      --bg-card-hover: #163060;
      --bg-input: #08152b;
      --bg-subtle: #132a54;

      --border-main: #1d3c6a;
      --border-subtle: #162e52;
      --border-focus: #00d8ff;
      --card-border: #1d3c6a;
      --card-header-bg: rgba(13, 31, 64, 0.65);
      --deck-bg: linear-gradient(135deg, rgba(23, 37, 65, 0.75) 0%, var(--bg-card) 100%);
      --th-bg: #0d1f40;
      --stat-top-border: rgba(0, 216, 255, 0.40);

      --text-bright: #f0fdfa;
      --text-main: #e0f2fe;
      --text-muted: #7dd3fc;
      --text-subtle: #38bdf8;

      --accent-primary: #00d8ff;
      --accent-cobalt: #0284c7;
      --accent-sky: #00d8ff;
      --accent-indigo: #818cf8;
      --accent-emerald: #10b981;
      --accent-emerald-bg: rgba(16, 185, 129, 0.18);
      --accent-amber: #f59e0b;
      --accent-amber-bg: rgba(245, 158, 11, 0.18);
      --accent-crimson: #ef4444;
      --accent-crimson-bg: rgba(239, 68, 68, 0.18);
      --accent-blue: #0ea5e9;

      --theme-active-border: #00d8ff;
      --theme-glow: rgba(0, 216, 255, 0.22);
      --theme-brand-pill: rgba(0, 216, 255, 0.16);
      --theme-brand-pill-text: #38bdf8;
      --theme-brand-pill-border: rgba(0, 216, 255, 0.35);

      --shadow-card: 0 4px 16px rgba(2, 8, 20, 0.50);
      --shadow-modal: 0 20px 50px rgba(2, 6, 15, 0.85), 0 0 0 1px var(--card-border);
    }

    /* THEME 2: CARBON BLACK (STEALTH OPERATIONS // VIBRANT TELEMETRY // NO COLORED BORDERS) */
    [data-theme="carbon-black"], [data-theme="stealth-ops"] {
      --bg-void: #09090b;
      --bg-base: #0d0d10;
      --bg-sidebar: #111114;
      --bg-header: #131317;
      --bg-card: #16161b;
      --bg-card-hover: #1e1e25;
      --bg-input: #121216;
      --bg-subtle: #1a1a20;

      --border-main: #27272e;
      --border-subtle: #1e1e24;
      --border-focus: #3b82f6;
      --card-border: #27272e;
      --card-header-bg: rgba(19, 19, 23, 0.70);
      --deck-bg: linear-gradient(135deg, rgba(28, 30, 40, 0.85) 0%, var(--bg-card) 100%);
      --th-bg: #141418;
      --stat-top-border: rgba(59, 130, 246, 0.35);

      --text-bright: #ffffff;
      --text-main: #e4e4e7;
      --text-muted: #a1a1aa;
      --text-subtle: #71717a;

      --accent-primary: #3b82f6;
      --accent-cobalt: #2563eb;
      --accent-sky: #38bdf8;
      --accent-indigo: #6366f1;
      --accent-emerald: #10b981;
      --accent-emerald-bg: rgba(16, 185, 129, 0.15);
      --accent-amber: #f59e0b;
      --accent-amber-bg: rgba(245, 158, 11, 0.15);
      --accent-crimson: #ef4444;
      --accent-crimson-bg: rgba(239, 68, 68, 0.15);
      --accent-blue: #3b82f6;

      --theme-active-border: #3b82f6;
      --theme-glow: rgba(59, 130, 246, 0.22);
      --theme-brand-pill: rgba(59, 130, 246, 0.18);
      --theme-brand-pill-text: #60a5fa;
      --theme-brand-pill-border: rgba(59, 130, 246, 0.35);

      --shadow-card: 0 4px 16px rgba(0, 0, 0, 0.50);
      --shadow-modal: 0 20px 50px rgba(0, 0, 0, 0.85), 0 0 0 1px var(--card-border);
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      background-color: var(--bg-void);
      color: var(--text-main);
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
      font-size: 13px;
      line-height: 1.5;
      height: 100vh;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    /* TOP STATUS BAR */
    .top-status-bar {
      height: 36px;
      background: var(--bg-header);
      border-bottom: 1px solid var(--border-main);
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 1.25rem;
      font-size: 0.72rem;
      color: var(--text-muted);
      user-select: none;
      flex-shrink: 0;
      z-index: 60;
    }
    .top-bar-left {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      font-weight: 600;
    }
    .top-status-indicator {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: var(--accent-emerald);
      box-shadow: 0 0 8px var(--accent-emerald);
    }
    .top-bar-right {
      display: flex;
      align-items: center;
      gap: 1.25rem;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.70rem;
    }

    /* APP LAYOUT */
    .app-layout {
      display: flex;
      flex: 1;
      height: calc(100vh - 36px);
      overflow: hidden;
      position: relative;
    }

    /* LEFT NAVIGATION SIDEBAR */
    .app-sidebar {
      width: 250px;
      min-width: 250px;
      max-width: 250px;
      background: var(--bg-sidebar);
      border-right: 1px solid var(--border-main);
      display: flex;
      flex-direction: column;
      user-select: none;
      height: 100%;
      transition: width 0.22s cubic-bezier(0.4, 0, 0.2, 1), min-width 0.22s cubic-bezier(0.4, 0, 0.2, 1), max-width 0.22s cubic-bezier(0.4, 0, 0.2, 1);
      position: relative;
      flex-shrink: 0;
      z-index: 50;
    }

    /* SIDEBAR COLLAPSED STATE */
    .app-sidebar.collapsed {
      width: 58px;
      min-width: 58px;
      max-width: 58px;
    }
    .app-sidebar.collapsed .brand-titles-wrap,
    .app-sidebar.collapsed .sidebar-range-pod,
    .app-sidebar.collapsed .nav-section-title,
    .app-sidebar.collapsed .nav-text,
    .app-sidebar.collapsed .nav-item-badge,
    .app-sidebar.collapsed #sidebarUserLabel,
    .app-sidebar.collapsed .footer-action-link-text {
      display: none !important;
    }
    .app-sidebar.collapsed .sidebar-brand-box {
      justify-content: center;
      padding: 0.85rem 0.5rem;
    }
    .app-sidebar.collapsed .nav-item {
      justify-content: center;
      padding: 0.65rem 0;
      margin: 0 0.35rem;
    }
    .app-sidebar.collapsed .sidebar-footer {
      flex-direction: column;
      padding: 0.65rem 0.25rem;
      gap: 0.5rem;
    }

    .sidebar-brand-box {
      padding: 1.10rem 1.25rem;
      display: flex;
      align-items: center;
      gap: 0.85rem;
      border-bottom: 1px solid var(--border-subtle);
      flex-shrink: 0;
    }
    .rf-brand-logo-img {
      width: 32px;
      height: 32px;
      object-fit: contain;
      flex-shrink: 0;
      transition: transform 0.18s ease;
      filter: drop-shadow(0 2px 6px rgba(0, 0, 0, 0.4));
    }
    .rf-brand-logo-img:hover {
      transform: translateY(-2px) scale(1.05);
    }
    .brand-titles-wrap {
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    .brand-title {
      font-size: 0.92rem;
      font-weight: 700;
      color: var(--text-bright);
      letter-spacing: -0.01em;
      white-space: nowrap;
    }
    .brand-subtitle {
      font-size: 0.67rem;
      color: var(--text-muted);
      white-space: nowrap;
    }

    /* Range Identity Pod in Sidebar */
    .sidebar-range-pod {
      margin: 0.85rem 0.85rem 0.4rem 0.85rem;
      background: var(--bg-card);
      border: 1px solid var(--card-border) !important;
      border-radius: 6px;
      padding: 0.65rem 0.75rem;
      display: flex;
      flex-direction: column;
      gap: 0.2rem;
      flex-shrink: 0;
      transition: all 0.2s ease;
    }
    .sidebar-range-pod:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
      border-color: rgba(56, 189, 248, 0.35) !important;
    }
    .sidebar-range-label {
      font-size: 0.62rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-subtle);
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .sidebar-range-name {
      font-size: 0.82rem;
      font-weight: 700;
      color: var(--text-bright);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .sidebar-range-meta {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.68rem;
      color: var(--text-muted);
    }

    /* Navigation Links Container */
    .sidebar-nav-list {
      padding: 0.5rem 0.65rem;
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
      flex: 1;
      min-height: 0;
      overflow-y: auto;
      overflow-x: hidden;
    }
    .sidebar-nav-list::-webkit-scrollbar {
      width: 4px;
    }
    .sidebar-nav-list::-webkit-scrollbar-thumb {
      background: var(--border-main);
      border-radius: 2px;
    }
    .nav-section-title {
      font-size: 0.62rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-subtle);
      padding: 0.5rem 0.5rem 0.25rem 0.5rem;
      white-space: nowrap;
    }
    .nav-item {
      display: flex;
      align-items: center;
      gap: 0.65rem;
      padding: 0.55rem 0.70rem;
      border-radius: 6px;
      color: var(--text-muted);
      text-decoration: none;
      font-size: 0.80rem;
      font-weight: 500;
      transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
      cursor: pointer;
      white-space: nowrap;
      position: relative;
    }
    .nav-item svg {
      flex-shrink: 0;
    }
    .nav-item:hover {
      background: var(--bg-card-hover);
      color: var(--text-bright);
      transform: translateY(-1.5px) translateX(2px);
    }
    .nav-item.active {
      background: rgba(37, 99, 235, 0.18);
      color: #38bdf8;
      font-weight: 600;
      border: 1px solid var(--border-main);
      border-left: 3px solid var(--theme-active-border) !important;
      box-shadow: 0 0 12px var(--theme-glow);
    }
    .brand-pill {
      display: inline-block;
      font-size: 0.60rem;
      font-weight: 700;
      letter-spacing: 0.05em;
      text-transform: uppercase;
      padding: 1px 6px;
      border-radius: 4px;
      background: var(--theme-brand-pill);
      color: var(--theme-brand-pill-text);
      border: 1px solid var(--theme-brand-pill-border);
      vertical-align: middle;
      margin-left: 0.35rem;
    }
    .nav-item-badge {
      margin-left: auto;
      font-size: 0.65rem;
      font-family: 'JetBrains Mono', monospace;
      padding: 1px 5px;
      border-radius: 4px;
      background: var(--bg-subtle);
      color: #38bdf8;
    }

    /* PINNED SIDEBAR FOOTER */
    .sidebar-footer {
      padding: 0.75rem 1rem;
      border-top: 1px solid var(--border-main);
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 0.75rem;
      color: var(--text-muted);
      background: var(--bg-sidebar);
      flex-shrink: 0;
      margin-top: auto;
      z-index: 10;
    }
    .footer-action-link {
      background: transparent;
      border: none;
      color: var(--text-muted);
      cursor: pointer;
      font-size: 0.72rem;
      text-decoration: none;
      padding: 3px 6px;
      border-radius: 4px;
      transition: transform 0.18s ease, color 0.15s ease, background 0.15s ease;
      display: inline-flex;
      align-items: center;
      gap: 0.3rem;
    }
    .footer-action-link:hover {
      color: #38bdf8;
      background: var(--bg-card-hover);
      transform: translateY(-1.5px);
    }

    /* RIGHT MAIN CONTENT VIEWPORT */
    .c2-main-viewport {
      flex: 1;
      min-width: 0;
      display: flex;
      flex-direction: column;
      overflow-y: auto;
      background: var(--bg-base);
      transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .c2-main-viewport::-webkit-scrollbar {
      width: 6px;
    }
    .c2-main-viewport::-webkit-scrollbar-thumb {
      background: var(--border-main);
      border-radius: 3px;
    }

    /* PAGE HEADER */
    .page-header {
      background: var(--bg-header);
      border-bottom: 1px solid var(--border-main);
      padding: 0.85rem 1.75rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      position: sticky;
      top: 0;
      z-index: 40;
      flex-shrink: 0;
    }
    .page-header-left {
      display: flex;
      align-items: center;
      gap: 0.85rem;
    }
    .sidebar-toggle-btn {
      background: var(--bg-card);
      border: 1px solid var(--card-border) !important;
      color: var(--text-bright);
      width: 32px;
      height: 32px;
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s ease, background-color 0.15s ease, color 0.15s ease;
    }
    .sidebar-toggle-btn:hover {
      background: var(--bg-card-hover);
      color: #38bdf8;
      transform: translateY(-2px);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
      border-color: rgba(56, 189, 248, 0.4) !important;
    }
    .page-title-wrap {
      display: flex;
      flex-direction: column;
      gap: 0.1rem;
    }
    .page-title {
      font-size: 1.15rem;
      font-weight: 700;
      color: var(--text-bright);
      letter-spacing: -0.02em;
    }
    .page-subtitle {
      font-size: 0.72rem;
      color: var(--text-muted);
    }
    .page-header-right {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }
    .header-util-btn {
      background: var(--bg-card);
      border: 1px solid var(--card-border) !important;
      color: var(--text-muted);
      padding: 0.35rem 0.65rem;
      border-radius: 6px;
      font-size: 0.72rem;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 0.4rem;
      transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s ease, background-color 0.15s ease, color 0.15s ease;
    }
    .header-util-btn:hover {
      background: var(--bg-card-hover);
      color: #38bdf8;
      transform: translateY(-2px);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
      border-color: rgba(56, 189, 248, 0.4) !important;
    }

    /* UNIVERSAL BUTTONS HOVER & TACTILE ELEVATION */
    button, .btn, .tab-btn, .filter-btn, .sidebar-toggle-btn, .header-util-btn, .action-btn, .ops-btn, .modal-close {
      display: inline-flex;
      align-items: center;
      gap: 0.45rem;
      padding: 0.45rem 0.85rem;
      border-radius: 6px;
      font-size: 0.76rem;
      font-weight: 600;
      cursor: pointer;
      border: 1px solid transparent;
      transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s ease, background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
      text-decoration: none;
      user-select: none;
    }
    button:hover, .btn:hover, .tab-btn:hover, .filter-btn:hover, .sidebar-toggle-btn:hover, .header-util-btn:hover, .action-btn:hover, .ops-btn:hover, .modal-close:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
    }
    button:active, .btn:active {
      transform: translateY(0);
    }
    .btn-primary {
      background: #10b981;
      color: #042f1f;
      border-color: #10b981;
    }
    .btn-primary:hover {
      background: #059669;
      border-color: #059669;
    }
    .btn-secondary {
      background: var(--bg-card);
      color: var(--text-main);
      border: 1px solid var(--card-border) !important;
    }
    .btn-secondary:hover {
      background: var(--bg-card-hover);
      color: #38bdf8;
      border-color: rgba(56, 189, 248, 0.4) !important;
    }
    .btn-danger {
      background: var(--accent-crimson);
      color: #ffffff;
      border-color: var(--accent-crimson);
    }
    .btn-danger:hover {
      background: #dc2626;
      border-color: #dc2626;
    }

    /* UNIVERSAL FIELDS HOVER & TACTILE FLOATING MICRO-ANIMATION */
    input, select, textarea, .form-input, .editor-textarea {
      background: var(--bg-input);
      border: 1px solid var(--border-main);
      color: var(--text-bright);
      border-radius: 6px;
      padding: 0.50rem 0.75rem;
      font-size: 0.80rem;
      font-family: inherit;
      transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.18s ease, border-color 0.15s ease, background-color 0.15s ease;
      outline: none;
    }
    input:hover, select:hover, textarea:hover, .form-input:hover, .editor-textarea:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
      border-color: var(--border-focus);
    }
    input:focus, select:focus, textarea:focus, .form-input:focus, .editor-textarea:focus {
      transform: translateY(-2px);
      box-shadow: 0 0 0 2px var(--theme-glow), 0 4px 16px rgba(0, 0, 0, 0.45);
      border-color: var(--border-focus);
    }
    .form-input-mono {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.78rem;
    }

    /* VIEW PANELS */
    .view-panel {
      display: none;
      padding: 1.75rem;
      flex-direction: column;
      gap: 1.5rem;
      max-width: 1680px;
      width: 100%;
      margin: 0 auto;
    }
    .view-panel.active {
      display: flex;
    }

    /* CARDS & BOXES WITH CLEAN REFINED SLATE BORDERS (NO HARSH COLORED BORDERS) */
    .card, .stat-card, .operations-deck-card, .editor-container, .modal-card, .terminal-container {
      background: var(--bg-card);
      border: 1px solid var(--card-border) !important;
      border-radius: 8px;
      box-shadow: var(--shadow-card);
      overflow: hidden;
      transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
    }
    .card:hover, .stat-card:hover, .operations-deck-card:hover {
      border-color: rgba(56, 189, 248, 0.35) !important;
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.40);
    }
    .card-header {
      padding: 0.90rem 1.25rem;
      border-bottom: 1px solid var(--border-subtle);
      display: flex;
      align-items: center;
      justify-content: space-between;
      background: var(--card-header-bg);
    }
    .card-title {
      font-size: 0.85rem;
      font-weight: 700;
      color: var(--text-bright);
      letter-spacing: -0.01em;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
    .card-body {
      padding: 1.25rem;
    }

    /* INTEGRATED OPERATIONS CONTROL DECK (RICH BLUE ACCENT GRADIENT) */
    .operations-deck-card {
      padding: 1.15rem 1.4rem;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1.5rem;
      flex-wrap: wrap;
      background: var(--deck-bg);
      border: 1px solid var(--card-border) !important;
    }
    .deck-left {
      display: flex;
      align-items: center;
      gap: 1.5rem;
      flex-wrap: wrap;
    }
    .deck-range-title-group {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }
    .deck-range-name {
      font-size: 1.10rem;
      font-weight: 700;
      color: var(--text-bright);
    }
    .deck-controls-meta {
      display: flex;
      align-items: center;
      gap: 0.65rem;
    }
    .deck-right {
      display: flex;
      align-items: center;
      gap: 0.65rem;
    }

    /* STATS GRID WITH COLOR ACCENTS */
    .stats-row {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 1rem;
    }
    .stat-card {
      padding: 1.1rem 1.25rem;
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
      border-top: 2px solid var(--stat-top-border) !important;
    }
    .stat-label {
      font-size: 0.68rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-subtle);
    }
    .stat-value {
      font-family: 'JetBrains Mono', monospace;
      font-size: 1.65rem;
      font-weight: 700;
      color: var(--text-bright);
      line-height: 1.2;
    }
    .stat-desc {
      font-size: 0.68rem;
      color: var(--text-muted);
    }

    /* DATA TABLES */
    .table-wrap {
      overflow-x: auto;
      width: 100%;
    }
    .table-wrap::-webkit-scrollbar {
      height: 6px;
    }
    .table-wrap::-webkit-scrollbar-thumb {
      background: var(--border-main);
      border-radius: 3px;
    }
    .data-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.78rem;
      text-align: left;
    }
    .data-table th {
      padding: 0.65rem 1rem;
      background: var(--th-bg);
      color: var(--text-muted);
      font-weight: 600;
      text-transform: uppercase;
      font-size: 0.65rem;
      letter-spacing: 0.05em;
      border-bottom: 1px solid var(--border-main);
      white-space: nowrap;
    }
    .data-table td {
      padding: 0.70rem 1rem;
      border-bottom: 1px solid var(--border-subtle);
      color: var(--text-main);
      vertical-align: middle;
    }
    .data-table tr:last-child td {
      border-bottom: none;
    }
    .data-table tbody tr:hover {
      background: var(--bg-card-hover);
    }

    /* STATUS BADGES */
    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      padding: 2px 7px;
      border-radius: 4px;
      font-size: 0.65rem;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }
    .status-running {
      background: var(--accent-emerald-bg);
      color: var(--accent-emerald);
      border: 1px solid rgba(16, 185, 129, 0.3);
    }
    .status-paused {
      background: rgba(56, 189, 248, 0.14);
      color: #38bdf8;
      border: 1px solid rgba(56, 189, 248, 0.3);
    }
    .status-stopped {
      background: var(--accent-crimson-bg);
      color: var(--accent-crimson);
      border: 1px solid rgba(239, 68, 68, 0.3);
    }
    .status-standby {
      background: var(--accent-amber-bg);
      color: var(--accent-amber);
      border: 1px solid rgba(245, 158, 11, 0.3);
    }

    /* TOOL BADGES (NETWORKING VS HOST UE CAPABILITIES) */
    .tool-badge {
      display: inline-flex;
      align-items: center;
      gap: 3px;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.64rem;
      font-weight: 500;
      padding: 2px 6px;
      border-radius: 4px;
      margin: 1px 2px;
      white-space: nowrap;
      transition: transform 0.15s ease;
    }
    .tool-badge:hover {
      transform: translateY(-1.5px);
    }
    .tool-badge-net {
      background: rgba(37, 99, 235, 0.16);
      color: #60a5fa;
      border: 1px solid rgba(59, 130, 246, 0.3);
    }
    .tool-badge-host {
      background: rgba(16, 185, 129, 0.16);
      color: #34d399;
      border: 1px solid rgba(16, 185, 129, 0.3);
    }
    .os-badge {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      padding: 2px 7px;
      border-radius: 4px;
      font-size: 0.68rem;
      font-weight: 600;
      background: var(--bg-subtle);
      border: 1px solid var(--border-subtle);
      color: var(--text-bright);
    }

    /* STEALTH SHELL BANNER (ELEGANT DEEP BLUE SAPPHIRE TINT) */
    .stealth-shell-banner {
      background: rgba(30, 58, 138, 0.16);
      border: 1px solid rgba(59, 130, 246, 0.30);
      border-radius: 6px;
      padding: 0.75rem 1rem;
      display: flex;
      align-items: center;
      gap: 0.75rem;
      font-size: 0.75rem;
      color: #bfdbfe;
    }
    .stealth-shell-banner svg {
      flex-shrink: 0;
      color: #38bdf8;
    }

    /* TERMINAL CONTAINER */
    .terminal-container {
      background: var(--bg-void);
      border: 1px solid var(--border-main) !important;
      border-radius: 6px;
      padding: 1rem;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.75rem;
      color: #cbd5e1;
      min-height: 220px;
      max-height: 380px;
      overflow-y: auto;
      white-space: pre-wrap;
      word-break: break-all;
    }
    .terminal-container::-webkit-scrollbar {
      width: 6px;
    }
    .terminal-container::-webkit-scrollbar-thumb {
      background: var(--border-main);
      border-radius: 3px;
    }
    .cmd-run-block {
      margin-bottom: 0.85rem;
      border-left: 2px solid #38bdf8;
      padding-left: 0.65rem;
    }
    .cmd-run-header {
      color: var(--text-muted);
      font-size: 0.70rem;
      margin-bottom: 0.25rem;
    }
    .cmd-run-body {
      color: #f8fafc;
    }

    /* FORM STYLES */
    .form-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 1.25rem;
    }
    .form-group {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .form-label {
      font-size: 0.72rem;
      font-weight: 600;
      color: var(--text-muted);
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }
    .form-help {
      font-size: 0.68rem;
      color: var(--text-subtle);
    }

    /* EDITORS */
    .editor-textarea {
      width: 100%;
      min-height: 240px;
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.80rem;
      line-height: 1.6;
      resize: vertical;
    }

    /* PERSONA CARDS CATALOG */
    .persona-catalog-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
      gap: 1rem;
    }
    .persona-profile-card {
      background: var(--bg-subtle);
      border: 1px solid var(--border-main);
      border-radius: 6px;
      padding: 1rem;
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      transition: all 0.2s ease;
    }
    .persona-profile-card:hover {
      transform: translateY(-2px);
      border-color: rgba(56, 189, 248, 0.4);
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
    }
    .persona-title-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .persona-title {
      font-weight: 700;
      font-size: 0.85rem;
      color: var(--text-bright);
    }

    /* MODAL */
    .modal-overlay {
      position: fixed;
      top: 0;
      left: 0;
      width: 100vw;
      height: 100vh;
      background: rgba(5, 10, 20, 0.80);
      backdrop-filter: blur(4px);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 100;
    }
    .modal-overlay.active {
      display: flex;
    }
    .modal-card {
      width: 540px;
      max-width: 90vw;
      background: var(--bg-card);
      box-shadow: var(--shadow-modal);
    }
    .modal-body {
      padding: 1.25rem;
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    .modal-footer {
      padding: 0.85rem 1.25rem;
      border-top: 1px solid var(--border-subtle);
      display: flex;
      align-items: center;
      justify-content: flex-end;
      gap: 0.65rem;
      background: var(--card-header-bg);
    }

    /* TOAST */
    .toast-popup {
      position: fixed;
      bottom: 24px;
      right: 24px;
      background: var(--bg-card);
      border: 1px solid var(--border-focus);
      color: var(--text-bright);
      padding: 0.65rem 1.15rem;
      border-radius: 6px;
      font-size: 0.78rem;
      font-weight: 500;
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
      opacity: 0;
      transform: translateY(10px);
      transition: all 0.2s ease;
      pointer-events: none;
      z-index: 200;
    }
    .toast-popup.show {
      opacity: 1;
      transform: translateY(0);
    }

    .copy-chip {
      cursor: pointer;
      border-bottom: 1px dashed var(--border-focus);
      color: #38bdf8;
      transition: color 0.15s ease;
    }
    .copy-chip:hover {
      color: var(--accent-emerald);
    }

    .mono {
      font-family: 'JetBrains Mono', monospace;
    }
    .text-muted {
      color: var(--text-muted);
    }
    .text-bright {
      color: var(--text-bright);
    }

    /* REAL-TIME TELEMETRY SPARKLINES */
    .sparkline-svg {
      width: 100%;
      height: 28px;
      margin-top: 6px;
      overflow: visible;
    }
    .sparkline-line-cpu {
      fill: none;
      stroke: #60a5fa;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
    .sparkline-line-ram {
      fill: none;
      stroke: #a78bfa;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }

    /* OPERATIONS DECK SESSION TICKER */
    .deck-session-ticker {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      background: var(--bg-input);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 0.35rem 0.65rem;
    }
    .ticker-pulse-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--accent-amber);
      display: inline-block;
      transition: all 0.25s ease;
    }
    .ticker-pulse-dot.running {
      background: var(--accent-emerald);
      box-shadow: 0 0 10px var(--accent-emerald);
      animation: tickerPulse 1.4s infinite ease-in-out;
    }
    .ticker-pulse-dot.stopped {
      background: var(--accent-crimson);
      box-shadow: 0 0 8px var(--accent-crimson);
    }
    @keyframes tickerPulse {
      0% { transform: scale(0.9); opacity: 0.7; }
      50% { transform: scale(1.25); opacity: 1; }
      100% { transform: scale(0.9); opacity: 0.7; }
    }

    /* ACTIVITY STREAM FILTER PILLS & SEARCH */
    .filter-pills-group {
      display: inline-flex;
      background: var(--bg-input);
      border: 1px solid var(--border-subtle);
      border-radius: 6px;
      padding: 2px;
      gap: 2px;
    }
    .filter-pill {
      padding: 2px 8px;
      font-size: 0.68rem;
      font-weight: 600;
      border: none;
      background: transparent;
      color: var(--text-muted);
      border-radius: 4px;
      cursor: pointer;
      transition: all 0.15s ease;
      user-select: none;
    }
    .filter-pill:hover {
      color: var(--text-bright);
      background: var(--bg-subtle);
    }
    .filter-pill.active {
      background: rgba(0, 216, 255, 0.18);
      color: #38bdf8;
      font-weight: 700;
    }
    .activity-search-input {
      padding: 0.28rem 0.55rem;
      font-size: 0.70rem;
      border-radius: 5px;
      width: 160px;
      background: var(--bg-input);
      border: 1px solid var(--border-subtle);
      color: var(--text-bright);
      outline: none;
      transition: all 0.18s ease;
    }
    .activity-search-input:focus {
      border-color: var(--border-focus);
      box-shadow: 0 0 0 2px var(--theme-glow);
    }
    .activity-live-badge {
      display: inline-flex;
      align-items: center;
      gap: 4px;
      font-size: 0.62rem;
      font-weight: 700;
      color: #10b981;
      letter-spacing: 0.05em;
      background: rgba(16, 185, 129, 0.14);
      padding: 2px 6px;
      border-radius: 4px;
      border: 1px solid rgba(16, 185, 129, 0.3);
    }
    .activity-live-dot {
      width: 5px;
      height: 5px;
      border-radius: 50%;
      background: #10b981;
      box-shadow: 0 0 6px #10b981;
      animation: tickerPulse 1.2s infinite;
    }

    /* TELEMETRY INSPECTOR BARS */
    .telemetry-bar-wrap {
      background: var(--bg-input);
      height: 8px;
      border-radius: 4px;
      overflow: hidden;
      margin-top: 4px;
      border: 1px solid var(--border-subtle);
      position: relative;
    }
    .telemetry-bar-fill {
      height: 100%;
      border-radius: 3px;
      transition: width 0.3s ease;
    }
    .shortcut-kbd {
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.70rem;
      font-weight: 600;
      background: var(--bg-input);
      border: 1px solid var(--border-main);
      padding: 2px 7px;
      border-radius: 4px;
      color: #38bdf8;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.25);
    }

    /* ==========================================================================
       RESPONSIVE DESIGN & VIEWPORT ADAPTATIONS
       Clean, fluid resizing across Desktop, Laptop, Tablet, and Mobile
       ========================================================================== */
    .sidebar-backdrop {
      display: none;
      position: fixed;
      inset: 0;
      background: rgba(4, 9, 20, 0.65);
      backdrop-filter: blur(4px);
      z-index: 85;
      opacity: 0;
      transition: opacity 0.22s ease;
      pointer-events: none;
    }
    .sidebar-backdrop.active {
      display: block;
      opacity: 1;
      pointer-events: auto;
    }

    @media (max-width: 1200px) {
      .operations-deck-card {
        padding: 1rem 1.25rem;
      }
      .view-panel {
        padding: 1.25rem;
      }
    }

    @media (max-width: 900px) {
      .app-sidebar {
        position: fixed;
        top: 36px;
        bottom: 0;
        left: 0;
        z-index: 95;
        width: 260px;
        min-width: 260px;
        max-width: 260px;
        box-shadow: 4px 0 24px rgba(0, 0, 0, 0.65);
        transform: translateX(0);
        transition: transform 0.24s cubic-bezier(0.4, 0, 0.2, 1);
      }
      .app-sidebar.collapsed {
        transform: translateX(-100%);
        width: 260px;
        min-width: 260px;
        max-width: 260px;
      }
      .app-sidebar.collapsed .brand-titles-wrap,
      .app-sidebar.collapsed .sidebar-range-pod,
      .app-sidebar.collapsed .nav-section-title,
      .app-sidebar.collapsed .nav-text,
      .app-sidebar.collapsed .nav-item-badge,
      .app-sidebar.collapsed #sidebarUserLabel,
      .app-sidebar.collapsed .footer-action-link-text {
        display: initial !important;
      }
      .app-sidebar.collapsed .sidebar-brand-box {
        justify-content: flex-start;
        padding: 1.10rem 1.25rem;
      }
      .app-sidebar.collapsed .nav-item {
        justify-content: flex-start;
        padding: 0.55rem 0.70rem;
        margin: 0;
      }
      .app-sidebar.collapsed .sidebar-footer {
        flex-direction: row;
        padding: 0.75rem 1rem;
      }
      .c2-main-viewport {
        width: 100%;
      }
      .deck-left, .deck-right {
        width: 100%;
      }
      .deck-right {
        justify-content: flex-start;
      }
      .top-bar-right span:nth-child(2) {
        display: none; /* Hide local clock on medium/small viewports */
      }
    }

    @media (max-width: 640px) {
      .top-status-bar {
        padding: 0 0.75rem;
        font-size: 0.68rem;
      }
      .top-bar-left span:nth-child(3),
      .top-bar-left #topActiveRangeLabel {
        display: none;
      }
      .page-header {
        padding: 0.65rem 0.85rem;
      }
      .page-title {
        font-size: 1.05rem;
      }
      .page-subtitle {
        font-size: 0.68rem;
      }
      .view-panel {
        padding: 0.85rem 0.65rem;
        gap: 1rem;
      }
      .card-header {
        padding: 0.75rem 0.95rem;
        flex-wrap: wrap;
        gap: 0.5rem;
      }
      .card-body {
        padding: 0.95rem;
      }
      .stats-row {
        grid-template-columns: repeat(2, 1fr);
        gap: 0.65rem;
      }
      .stat-card {
        padding: 0.75rem 0.85rem;
      }
      .stat-value {
        font-size: 1.35rem;
      }
      .operations-deck-card {
        padding: 0.85rem 0.95rem;
        gap: 1rem;
      }
      .deck-range-title-group {
        flex-wrap: wrap;
      }
      .deck-controls-meta {
        flex-direction: column;
        align-items: flex-start;
        width: 100%;
        gap: 0.35rem;
      }
      .deck-controls-meta select {
        width: 100% !important;
      }
      .deck-right {
        flex-wrap: wrap;
        width: 100%;
      }
      .deck-right button {
        flex: 1 1 calc(33.333% - 0.5rem);
        min-width: 90px;
        justify-content: center;
        padding: 0.45rem 0.5rem;
        font-size: 0.72rem;
      }
      .modal-card {
        width: 95vw;
        max-width: 95vw;
        border-radius: 8px;
      }
      .toast-popup {
        left: 16px;
        right: 16px;
        bottom: 16px;
        text-align: center;
      }
    }

    @media (max-width: 420px) {
      .stats-row {
        grid-template-columns: 1fr;
      }
      .deck-right button {
        flex: 1 1 100%;
      }
    }
  </style>
</head>
<body>

  <!-- TOP MINIMAL STATUS BAR -->
  <header class="top-status-bar">
    <div class="top-bar-left">
      <span class="top-status-indicator"></span>
      <span style="letter-spacing:0.04em;">RANGEFORGE CONTROLLER</span>
      <span style="color:#38bdf8;">|</span>
      <span style="color:var(--text-subtle);" id="topActiveRangeLabel">Cyber Range</span>
    </div>
    <div class="top-bar-right">
      <span>UTC: <strong id="utc-clock" style="color:var(--text-bright);">--:--:--</strong></span>
      <span>LOCAL: <strong id="loc-clock" style="color:var(--text-bright);">--:--:--</strong></span>
      <span style="color:#38bdf8;">HTTPS (PORT 8443)</span>
    </div>
  </header>

  <div class="app-layout">
    <!-- Mobile Sidebar Backdrop -->
    <div class="sidebar-backdrop" id="sidebarBackdrop" onclick="closeSidebarMobile()"></div>

    <!-- LEFT NAVIGATION SIDEBAR -->
    <aside class="app-sidebar" id="appSidebar">
      <!-- Brand Box -->
      <div class="sidebar-brand-box">
        <div class="brand-titles-wrap">
          <div class="brand-title">RANGEFORGE <span class="brand-pill">UE</span></div>
          <div class="brand-subtitle">User Emulation Platform</div>
        </div>
      </div>

      <!-- Active Range Status Pod -->
      <div class="sidebar-range-pod">
        <div class="sidebar-range-label">
          <span>Active Range</span>
          <span class="status-badge status-standby" id="sbRangeBadge">STANDBY</span>
        </div>
        <div class="sidebar-range-name" id="sbRangeName">Cyber Range</div>
        <div class="sidebar-range-meta" id="sbRangeCidr">10.0.0.0/16</div>
      </div>

      <!-- NAVIGATION LINKS -->
      <nav class="sidebar-nav-list">
        <div class="nav-section-title">Navigation</div>
        <a href="/" class="nav-item active" id="nav-overview" onclick="return navigatePage('/', event)" title="Overview">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
          <span class="nav-text">Overview</span>
        </a>
        <a href="/fleet" class="nav-item" id="nav-fleet" onclick="return navigatePage('/fleet', event)" title="Connected Fleet & Personas">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
          <span class="nav-text">Fleet &amp; Personas</span>
          <span class="nav-item-badge" id="navFleetCount">0</span>
        </a>
        <a href="/emulation" class="nav-item" id="nav-emulation" onclick="return navigatePage('/emulation', event)" title="User Emulation Configuration">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
          <span class="nav-text">Emulation Engine</span>
        </a>
        <a href="/corpus" class="nav-item" id="nav-corpus" onclick="return navigatePage('/corpus', event)" title="Web Corpus & Wordlists">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
          <span class="nav-text">Web Corpus &amp; Targets</span>
        </a>
        <a href="/topology" class="nav-item" id="nav-topology" onclick="return navigatePage('/topology', event)" title="Network Topology">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 2 7 12 12 22 7 12 2"></polygon><polyline points="2 17 12 22 22 17"></polyline><polyline points="2 12 12 17 22 12"></polyline></svg>
          <span class="nav-text">Network Topology</span>
        </a>
        <a href="/activity" class="nav-item" id="nav-activity" onclick="return navigatePage('/activity', event)" title="Activity & Audit Logs">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline></svg>
          <span class="nav-text">Activity &amp; Audit Logs</span>
          <span class="nav-item-badge" id="navEventCount">0</span>
        </a>
      </nav>

      <!-- PINNED SIDEBAR FOOTER -->
      <div class="sidebar-footer">
        <span id="sidebarUserLabel" style="font-weight:600; color:var(--text-bright);">admin</span>
        <div style="display:flex; gap:0.25rem;">
          <button class="footer-action-link" onclick="toggleTheme()" title="Toggle Theme (Ocean Sapphire Cobalt / Carbon Black)">
            <span class="footer-action-link-text">Theme</span>
          </button>
          <button class="footer-action-link" onclick="openAdminModal()" title="Change Admin Password">
            <span class="footer-action-link-text">Password</span>
          </button>
          <button class="footer-action-link" onclick="handleLogout()" style="color:var(--accent-crimson);" title="Sign Out">
            <span class="footer-action-link-text">Sign Out</span>
          </button>
        </div>
      </div>
    </aside>

    <!-- RIGHT MAIN CONTENT VIEWPORT -->
    <main class="c2-main-viewport" id="mainViewport">
      <!-- TOP PAGE HEADER -->
      <header class="page-header">
        <div class="page-header-left">
          <button class="sidebar-toggle-btn" id="sidebarToggleBtn" onclick="toggleSidebar()" title="Toggle Left Sidebar [Ctrl+B]">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>
          </button>
          <div class="page-title-wrap">
            <h1 class="page-title" id="mainHeaderTitle">Overview Console</h1>
            <div class="page-subtitle" id="mainHeaderSubtitle">Operational Situational Awareness</div>
          </div>
        </div>

        <div class="page-header-right">
          <span class="status-badge status-standby" id="headerStatusBadge">STANDBY</span>
          <button class="header-util-btn" onclick="openExecutiveReportModal()" title="Export Executive Range Brief">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>
            Export Brief
          </button>
          <button class="header-util-btn" onclick="toggleTheme()" title="Toggle Theme (Ocean Sapphire Cobalt / Carbon Black)">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
            Theme
          </button>
          <button class="header-util-btn" onclick="openShortcutsModal()" title="Keyboard Shortcuts [?]">
            <kbd class="shortcut-kbd" style="font-size:0.65rem; padding:1px 5px;">?</kbd>
          </button>
        </div>
      </header>

      <!-- VIEW 1: OVERVIEW -->
      <section class="view-panel active" id="view-overview">
        <!-- INTEGRATED OPERATIONS CONTROL DECK -->
        <div class="operations-deck-card">
          <div class="deck-left">
            <div class="deck-range-title-group">
              <span class="status-badge status-standby" id="deckRangeBadge">STANDBY</span>
              <span class="deck-range-name" id="deckRangeName">Cyber Range</span>
              <span class="mono text-muted" id="deckRangeCidr" style="font-size:0.75rem;">10.0.0.0/16</span>
            </div>
            <div class="deck-controls-meta">
              <label class="form-label" style="margin:0;">Intensity Profile:</label>
              <select class="form-input form-input-mono" id="deckIntensitySelect" style="width:auto; padding:0.35rem 0.65rem; font-size:0.75rem;" onchange="setRangeIntensity(this.value)">
                <option value="Low (0.25x)">Low (0.25x) - Background Traffic</option>
                <option value="Medium (1.0x)" selected>Medium (1.0x) - Standard Operations</option>
                <option value="High (3.0x)">High (3.0x) - Active Training Shift</option>
                <option value="Maximum (5.0x)">Maximum (5.0x) - Peak Surge Exercise</option>
              </select>
              <div class="deck-session-ticker" id="deckSessionTicker" title="Active Cyber Range Session Telemetry">
                <span class="ticker-pulse-dot" id="deckPulseDot"></span>
                <span class="mono" id="deckSessionTime" style="font-size:0.70rem; color:var(--text-bright);">SESSION: 00:00:00</span>
                <span style="color:var(--text-subtle); opacity:0.6;">|</span>
                <span class="mono" id="deckActionCount" style="font-size:0.70rem; color:#38bdf8;">THROUGHPUT: 0 OPS</span>
              </div>
            </div>
          </div>
          <div class="deck-right">
            <button class="btn btn-primary" id="btnStart" onclick="setRangeState('running')">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
              Start Emulation
            </button>
            <button class="btn btn-secondary" id="btnPause" onclick="setRangeState('paused')">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"></rect><rect x="14" y="4" width="4" height="16"></rect></svg>
              Pause
            </button>
            <button class="btn btn-danger" id="btnStop" onclick="setRangeState('stopped')">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2"></rect></svg>
              Stop Emulation
            </button>
          </div>
        </div>

        <!-- STATS ROW WITH RICH VIBRANT METRIC COLORS -->
        <div class="stats-row">
          <div class="stat-card">
            <span class="stat-label">Operational State</span>
            <span class="stat-value" id="statOpState" style="font-size:1.15rem; color:#f59e0b;">STANDBY</span>
            <span class="stat-desc">Range execution status</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Active Endpoints</span>
            <span class="stat-value" id="statEndpoints" style="color:#38bdf8;">0 / 0</span>
            <span class="stat-desc">Online host agents</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Traffic Intensity</span>
            <span class="stat-value" id="statIntensity" style="color:#f59e0b;">Medium</span>
            <span class="stat-desc">Global noise multiplier</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Fleet Avg CPU</span>
            <span class="stat-value" id="statCpu" style="color:#60a5fa;">0%</span>
            <span class="stat-desc">Aggregate endpoint load</span>
            <svg id="sparklineCpu" class="sparkline-svg" viewBox="0 0 120 28" preserveAspectRatio="none">
              <defs>
                <linearGradient id="cpuGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.45"/>
                  <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.0"/>
                </linearGradient>
              </defs>
              <polygon id="sparklineCpuPoly" points="0,28 120,28" fill="url(#cpuGrad)"/>
              <polyline id="sparklineCpuLine" class="sparkline-line-cpu" points="0,28 120,28"/>
            </svg>
          </div>
          <div class="stat-card">
            <span class="stat-label">Fleet Avg RAM</span>
            <span class="stat-value" id="statRam" style="color:#a78bfa;">0%</span>
            <span class="stat-desc">Memory consumption</span>
            <svg id="sparklineRam" class="sparkline-svg" viewBox="0 0 120 28" preserveAspectRatio="none">
              <defs>
                <linearGradient id="ramGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#a78bfa" stop-opacity="0.45"/>
                  <stop offset="100%" stop-color="#a78bfa" stop-opacity="0.0"/>
                </linearGradient>
              </defs>
              <polygon id="sparklineRamPoly" points="0,28 120,28" fill="url(#ramGrad)"/>
              <polyline id="sparklineRamLine" class="sparkline-line-ram" points="0,28 120,28"/>
            </svg>
          </div>
          <div class="stat-card">
            <span class="stat-label">Wordlist Files</span>
            <span class="stat-value" id="statFilesCreated" style="color:#34d399;">0</span>
            <span class="stat-desc">Created (<span id="statFilesDeleted">0</span> cleaned)</span>
          </div>
        </div>

        <!-- CONNECTED AGENT FLEET TABLE (OVERVIEW SUMMARY) -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">CONNECTED AGENT FLEET &amp; HOST INVENTORY TELEMETRY</span>
            <a href="/fleet" class="btn btn-secondary" style="padding:0.35rem 0.65rem; font-size:0.72rem;" onclick="return navigatePage('/fleet', event)">Manage Personas &amp; Commands</a>
          </div>
          <div class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Host Endpoint</th>
                  <th>Exact OS / Build</th>
                  <th>Network IP</th>
                  <th>Assigned Persona</th>
                  <th>CPU Usage</th>
                  <th>RAM Usage</th>
                  <th>Status</th>
                  <th style="text-align:right;">Actions</th>
                </tr>
              </thead>
              <tbody id="overviewFleetTableBody">
                <tr>
                  <td colspan="8" style="text-align:center; color:var(--text-muted); padding:2rem;">No agent hosts registered on this range yet.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- RECENT EMULATION EVENTS -->
        <div class="card">
          <div class="card-header" style="flex-wrap:wrap; gap:0.65rem;">
            <div style="display:flex; align-items:center; gap:0.75rem;">
              <span class="card-title">RECENT RANGE ACTIVITY STREAM</span>
              <span class="activity-live-badge" title="Live Auto-Polling Active"><span class="activity-live-dot"></span> LIVE</span>
            </div>
            <div style="display:flex; align-items:center; gap:0.5rem; flex-wrap:wrap;">
              <div class="filter-pills-group" id="overviewFilterPills">
                <button class="filter-pill active" onclick="setOverviewFilter('ALL', this)">ALL</button>
                <button class="filter-pill" onclick="setOverviewFilter('HTTP', this)">HTTP</button>
                <button class="filter-pill" onclick="setOverviewFilter('SMB', this)">SMB</button>
                <button class="filter-pill" onclick="setOverviewFilter('FILE', this)">FILE</button>
                <button class="filter-pill" onclick="setOverviewFilter('ICMP', this)">ICMP</button>
                <button class="filter-pill" onclick="setOverviewFilter('CMD', this)">CMD</button>
              </div>
              <input type="text" class="activity-search-input" id="overviewSearchInput" placeholder="Filter stream..." oninput="handleOverviewSearch(this.value)">
              <a href="/activity" class="btn btn-secondary" style="padding:0.30rem 0.60rem; font-size:0.70rem;" onclick="return navigatePage('/activity', event)">Full Log</a>
            </div>
          </div>
          <div class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Time (UTC)</th>
                  <th>Host</th>
                  <th>Protocol</th>
                  <th>Action Summary</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody id="overviewEventsBody">
                <tr>
                  <td colspan="5" style="text-align:center; color:var(--text-muted); padding:2rem;">Awaiting emulation activity events...</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <!-- VIEW 2: CONNECTED FLEET & PERSONA MANAGEMENT -->
      <section class="view-panel" id="view-fleet">
        <!-- 1. Connected Agent Fleet & Live Persona Management -->
        <div class="card">
          <div class="card-header">
            <div style="display:flex; flex-direction:column; gap:0.15rem;">
              <span class="card-title">CONNECTED AGENT FLEET &amp; LIVE PERSONA ASSIGNMENT</span>
              <span style="font-size:0.70rem; color:var(--text-muted);">Exact OS telemetry, detected UE traffic generation capabilities, and live persona assignment. Personas persist across restarts.</span>
            </div>
            <div style="display:flex; align-items:center; gap:0.75rem;">
              <span style="font-size:0.72rem; color:var(--text-muted);">Set All Agents:</span>
              <select class="form-input form-input-mono" id="bulkPersonaSelect" style="padding:0.35rem 0.65rem; font-size:0.75rem;">
                <option value="office_worker">Office Worker</option>
                <option value="developer">Developer</option>
                <option value="sysadmin">System Administrator</option>
                <option value="finance">Finance Specialist</option>
                <option value="executive">Corporate Executive</option>
                <option value="hr_specialist">HR Specialist</option>
                <option value="scada_operator">SCADA / ICS Operator</option>
              </select>
              <button class="btn btn-secondary" style="padding:0.35rem 0.65rem; font-size:0.72rem;" onclick="applyBulkPersona()">Apply Fleet-Wide</button>
            </div>
          </div>
          <div class="table-wrap">
            <table class="data-table" id="fleetDetailTable">
              <thead>
                <tr>
                  <th style="width:16%;">Host Endpoint</th>
                  <th style="width:20%;">Exact Operating System</th>
                  <th style="width:12%;">Network IP</th>
                  <th style="width:24%;">Detected UE Traffic Tools</th>
                  <th style="width:16%;">Assigned Persona</th>
                  <th style="width:6%;">Status</th>
                  <th style="width:6%; text-align:right;">Actions</th>
                </tr>
              </thead>
              <tbody id="fleetDetailTableBody">
                <tr>
                  <td colspan="7" style="text-align:center; color:var(--text-muted); padding:2rem;">No agent hosts registered on this range yet.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 2. FLEET REMOTE COMMAND EXECUTION CONSOLE -->
        <div class="card">
          <div class="card-header">
            <div style="display:flex; flex-direction:column; gap:0.15rem;">
              <span class="card-title">FLEET REMOTE COMMAND CONSOLE</span>
              <span style="font-size:0.70rem; color:var(--text-muted);">Execute authorized administrative and simulation commands across all hosts simultaneously or targeted endpoints.</span>
            </div>
            <span class="status-badge status-running" style="font-family:'JetBrains Mono', monospace;">STEALTH SHELL ACTIVE</span>
          </div>
          <div class="card-body" style="display:flex; flex-direction:column; gap:1rem;">
            <!-- Stealth History Suppression Banner -->
            <div class="stealth-shell-banner">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
              <div>
                <strong>Shell History Suppression Active:</strong> Commands executed through this manager inject <code class="mono">HISTFILE=/dev/null</code>, <code class="mono">HISTSIZE=0</code>, and PowerShell <code class="mono">-NoProfile</code> so commands are <strong>never written to user shell histories</strong> and will not confuse exercise participants.
              </div>
            </div>

            <!-- Command Dispatch Bar -->
            <div style="display:flex; gap:0.75rem; flex-wrap:wrap;">
              <div style="flex:0 0 260px;">
                <label class="form-label">Target Scope</label>
                <select class="form-input form-input-mono" id="fleetCmdTarget" style="width:100%;">
                  <option value="ALL">⚡ ALL HOSTS (Fleet-Wide Execution)</option>
                  <option value="ALL_WINDOWS">🪟 All Windows Hosts</option>
                  <option value="ALL_LINUX">🐧 All Linux Hosts</option>
                </select>
              </div>
              <div style="flex:1; min-width:300px;">
                <label class="form-label">Command Line</label>
                <div style="display:flex; gap:0.5rem;">
                  <input type="text" class="form-input form-input-mono" id="fleetCmdInput" placeholder="e.g. whoami, hostname, ipconfig, netstat -ano, systeminfo, uname -a..." style="flex:1;" onkeydown="if(event.key==='Enter') executeFleetCommand()">
                  <button class="btn btn-primary" onclick="executeFleetCommand()">Run Command</button>
                </div>
              </div>
            </div>

            <!-- Quick Command Presets -->
            <div style="display:flex; align-items:center; gap:0.5rem; flex-wrap:wrap;">
              <span class="form-label" style="margin:0;">Quick Presets:</span>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="setFleetCmd('whoami')">whoami</button>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="setFleetCmd('hostname')">hostname</button>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="setFleetCmd('ipconfig')">ipconfig</button>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="setFleetCmd('netstat -ano')">netstat -ano</button>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="setFleetCmd('systeminfo')">systeminfo</button>
              <button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.70rem; margin-left:auto;" onclick="clearFleetTerminal()">Clear Console</button>
            </div>

            <!-- Execution Results Terminal -->
            <div class="terminal-container" id="fleetCmdResults">
              <div style="color:var(--text-muted); font-size:0.72rem;">[RANGEFORGE STEALTH SHELL READY] Select target host(s), enter command, and execute. Zero footprint in user command history.</div>
            </div>
          </div>
        </div>

        <!-- 3. Persona Profiles Catalog Reference -->
        <div class="card">
          <div class="card-header">
            <span class="card-title">PERSONA CAPABILITY PROFILES CATALOG</span>
          </div>
          <div class="card-body">
            <div class="persona-catalog-grid">
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">Office Worker</span>
                  <span class="status-badge status-running">Standard</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Generates typical corporate HTTP/HTTPS browsing, intranet portal traffic, standard Word/Excel document creation and modification.</div>
              </div>
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">Developer</span>
                  <span class="status-badge status-running">Technical</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Simulates Git clone/pull/push operations, code repository browsing, API endpoint queries, and developer workstation filesystem activity.</div>
              </div>
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">System Administrator</span>
                  <span class="status-badge status-running">Privileged</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Generates SMB network share exploration, DNS lookups, administrative PowerShell queries, server ping health checks, and service telemetry.</div>
              </div>
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">Finance Specialist</span>
                  <span class="status-badge status-running">Enterprise</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Generates accounting portal sessions, secure HTTPS spreadsheet uploads, financial report text file manipulation, and database queries.</div>
              </div>
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">Corporate Executive</span>
                  <span class="status-badge status-running">Management</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Low-frequency, bursty executive web browsing, corporate dashboard reviews, confidential memo file creations, and email emulation.</div>
              </div>
              <div class="persona-profile-card">
                <div class="persona-title-row">
                  <span class="persona-title">SCADA / ICS Operator</span>
                  <span class="status-badge status-running">Industrial</span>
                </div>
                <div style="font-size:0.72rem; color:var(--text-muted);">Simulates field telemetry polling, periodic sensor readouts, industrial controller heartbeat pings, and operational technology traffic.</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- VIEW: USER EMULATION CONFIGURATION -->
      <section class="view-panel" id="view-emulation">
        <!-- 0. Blue Team Noise Posture & Binary Polling Control -->
        <div class="card" id="cardNoiseControl">
          <div class="card-header">
            <div>
              <span class="card-title">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
                Blue Team Noise &amp; Binary Polling Control
              </span>
              <div style="font-size:0.70rem; color:var(--text-muted); margin-top:2px;">Regulate host binary execution rates and network beaconing to prevent Sysmon / EDR alert spam. Confines file creation strictly to Documents, Downloads, and Desktop with zero simulation footprints.</div>
            </div>
            <div style="display:flex; gap:0.5rem; align-items:center;">
              <button class="btn btn-secondary" onclick="purgeAllArtifactsNow()" style="color:#ef4444; border-color:rgba(239,68,68,0.3);" title="Instantly purge all synthetic files from endpoints">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                Purge All UE Artifacts Now
              </button>
              <button class="btn btn-primary" onclick="saveNoiseControl()">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                Save Noise Pacing
              </button>
            </div>
          </div>
          <div class="card-body">
            <!-- Presets row -->
            <div style="margin-bottom:1.25rem;">
              <div style="font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:0.5rem;">Blue Team Noise Posture Presets:</div>
              <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(240px, 1fr)); gap:0.75rem;">
                <div class="noise-preset-box" id="presetStealth" onclick="setNoisePreset('stealth')" style="background:var(--bg-subtle); border:1px solid var(--border-subtle); border-radius:6px; padding:0.85rem; cursor:pointer; transition:all 0.2s ease;">
                  <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:4px;">
                    <span style="font-weight:700; font-size:0.78rem; color:#38bdf8; display:flex; align-items:center; gap:0.3rem;">
                      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>
                      Stealth Mode (EDR Quiet)
                    </span>
                    <span class="badge" style="background:rgba(56,189,248,0.15); color:#38bdf8; font-size:0.65rem;">Recommended</span>
                  </div>
                  <div style="font-size:0.68rem; color:var(--text-muted); line-height:1.3;">Suppresses frequent binary spawns (300s pacing). Spaces beaconing to 30s so trainees can isolate red team activity without Sysmon noise fatigue.</div>
                </div>

                <div class="noise-preset-box" id="presetBalanced" onclick="setNoisePreset('balanced')" style="background:var(--bg-subtle); border:1px solid var(--border-subtle); border-radius:6px; padding:0.85rem; cursor:pointer; transition:all 0.2s ease;">
                  <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:4px;">
                    <span style="font-weight:700; font-size:0.78rem; color:#10b981; display:flex; align-items:center; gap:0.3rem;">
                      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                      Balanced Mode
                    </span>
                    <span class="badge" style="background:rgba(16,185,129,0.15); color:#10b981; font-size:0.65rem;">Standard</span>
                  </div>
                  <div style="font-size:0.68rem; color:var(--text-muted); line-height:1.3;">Realistic employee operating pace (60s binary cadence, 10s beaconing). Natural balance of host and network telemetry.</div>
                </div>

                <div class="noise-preset-box" id="presetActive" onclick="setNoisePreset('active')" style="background:var(--bg-subtle); border:1px solid var(--border-subtle); border-radius:6px; padding:0.85rem; cursor:pointer; transition:all 0.2s ease;">
                  <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:4px;">
                    <span style="font-weight:700; font-size:0.78rem; color:#f59e0b; display:flex; align-items:center; gap:0.3rem;">
                      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
                      Active Training Mode
                    </span>
                    <span class="badge" style="background:rgba(245,158,11,0.15); color:#f59e0b; font-size:0.65rem;">High Telemetry</span>
                  </div>
                  <div style="font-size:0.68rem; color:var(--text-muted); line-height:1.3;">Rapid command execution (20s interval, 5s beaconing) for high-stress telemetry ingestion and SIEM load testing.</div>
                </div>
              </div>
            </div>

            <!-- Fine-grained controls -->
            <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(220px, 1fr)); gap:1rem; padding-top:1rem; border-top:1px solid var(--border-subtle);">
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Host Binary Execution Interval (seconds)</label>
                <input type="number" id="ncToolInterval" min="10" max="600" value="60" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Higher intervals keep Event ID 1 / 4688 logs quiet</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Fleet Beacon / Heartbeat Cadence (seconds)</label>
                <input type="number" id="ncHeartbeatSec" min="3" max="120" value="5" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Frequency of agent check-ins to manager</span>
              </div>
              <div style="display:flex; flex-direction:column; justify-content:center;">
                <label style="display:flex; align-items:center; gap:0.5rem; font-size:0.74rem; font-weight:600; color:var(--text-bright); cursor:pointer;">
                  <input type="checkbox" id="ncPauseTools">
                  Mute Host Binary Execution (EDR Quiet)
                </label>
                <span style="font-size:0.65rem; color:var(--text-subtle); margin-top:3px; margin-left:1.5rem;">Completely halts process spawning while keeping benign file work running.</span>
              </div>
            </div>

            <!-- Zero-footprint assurance banner -->
            <div style="margin-top:1rem; background:rgba(16,185,129,0.06); border:1px solid rgba(16,185,129,0.25); border-radius:6px; padding:0.65rem 0.85rem; display:flex; align-items:center; justify-content:space-between; flex-wrap:wrap; gap:0.5rem;">
              <div style="display:flex; align-items:center; gap:0.5rem;">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
                <span style="font-size:0.72rem; color:var(--text-bright); font-weight:600;">Zero-Footprint Storage Policy Active:</span>
                <span style="font-size:0.70rem; color:var(--text-muted);">Files strictly confined to <code style="color:#38bdf8;">Documents</code>, <code style="color:#38bdf8;">Downloads</code> &amp; <code style="color:#38bdf8;">Desktop</code>. Zero tool/brand artifacts on endpoints.</span>
              </div>
              <span class="badge" style="background:rgba(16,185,129,0.2); color:#10b981; font-size:0.68rem; font-weight:600;">Full Auto-Purge on Exit</span>
            </div>
          </div>
        </div>

        <!-- 1. Global User Emulation Dynamics (Direct UI for ranges.json) -->
        <div class="card">
          <div class="card-header">
            <div>
              <span class="card-title">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                Global Emulation Dynamics &amp; Realism Pacing
              </span>
              <div style="font-size:0.70rem; color:var(--text-muted); margin-top:2px;">Configures user dwell times, request rates, network jitter, and egress policies directly via UI without touching ranges.json.</div>
            </div>
            <button class="btn btn-primary" onclick="saveGlobalEmulationConfig()">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
              Save &amp; Apply Emulation Dynamics
            </button>
          </div>
          <div class="card-body">
            <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(220px, 1fr)); gap:1rem;">
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Traffic Intensity Profile</label>
                <select id="cfgIntensity" style="width:100%;">
                  <option value="Low (0.25x)">Low (0.25x) - Background Noise</option>
                  <option value="Medium (1.0x)" selected>Medium (1.0x) - Standard Business</option>
                  <option value="High (2.5x)">High (2.5x) - Active Operations</option>
                  <option value="Aggressive (5.0x)">Aggressive (5.0x) - Stress Testing</option>
                </select>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Operating Mode</label>
                <select id="cfgOperatingMode" style="width:100%;">
                  <option value="Autonomous Emulation" selected>Autonomous Emulation</option>
                  <option value="Guided Exercise Scenario">Guided Exercise Scenario</option>
                  <option value="Interactive Range Operations">Interactive Range Operations</option>
                </select>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Egress Routing Policy</label>
                <select id="cfgEgressPolicy" style="width:100%;">
                  <option value="Simulated Internet Proxy Egress" selected>Simulated Internet Proxy Egress</option>
                  <option value="Direct Enclave Gateway Routing">Direct Enclave Gateway Routing</option>
                  <option value="Isolated Intranet Enclave Only">Isolated Intranet Enclave Only</option>
                </select>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Simulated Packet Loss (%)</label>
                <input type="number" id="cfgPacketLoss" step="0.1" min="0" max="25" value="0.0" style="width:100%;">
              </div>
            </div>

            <div style="margin-top:1rem; padding-top:1rem; border-top:1px solid var(--border-subtle); display:grid; grid-template-columns:repeat(auto-fit, minmax(220px, 1fr)); gap:1rem;">
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Min Dwell Time (seconds)</label>
                <input type="number" id="cfgDwellMin" min="1" max="300" value="5" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Pause between consecutive user actions</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Max Dwell Time (seconds)</label>
                <input type="number" id="cfgDwellMax" min="1" max="600" value="25" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Upper threshold for human-like pauses</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Min Web Requests (req/min)</label>
                <input type="number" id="cfgWebReqMin" min="1" max="100" value="4" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Minimum browsing frequency per agent</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Max Web Requests (req/min)</label>
                <input type="number" id="cfgWebReqMax" min="1" max="500" value="16" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Peak burst browsing requests</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Network Latency Injection (ms)</label>
                <input type="number" id="cfgLatencyMs" min="0" max="1000" value="15" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Simulated round-trip latency</span>
              </div>
              <div>
                <label class="form-label" style="display:block; font-size:0.72rem; font-weight:600; color:var(--text-muted); margin-bottom:4px;">Network Jitter (%)</label>
                <input type="number" id="cfgJitterPct" min="0" max="100" value="25" style="width:100%;">
                <span class="form-help" style="font-size:0.65rem; color:var(--text-subtle);">Random variance applied to timing</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Persona Emulation Behavior Studio (Direct UI for Profiles) -->
        <div class="card">
          <div class="card-header">
            <div style="display:flex; align-items:center; gap:1rem;">
              <span class="card-title">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                Persona Behavior Studio
              </span>
              <div style="display:flex; align-items:center; gap:0.5rem;">
                <span style="font-size:0.70rem; color:var(--text-muted); font-weight:600;">Active Persona:</span>
                <select id="personaStudioSelect" onchange="loadSelectedPersonaStudio()" style="padding:0.3rem 0.6rem; font-size:0.76rem;">
                  <option value="office_worker">Office Worker</option>
                  <option value="developer">Developer</option>
                  <option value="sysadmin">Sysadmin</option>
                  <option value="executive">Executive</option>
                  <option value="finance">Finance Specialist</option>
                  <option value="scada_operator">SCADA / ICS Operator</option>
                </select>
              </div>
            </div>
            <button class="btn btn-primary" onclick="savePersonaStudioConfig()">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
              Save Persona Profile
            </button>
          </div>
          <div class="card-body">
            <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(320px, 1fr)); gap:1.25rem;">
              <!-- Module A: Web Browsing Activity -->
              <div style="background:var(--bg-subtle); padding:1rem; border-radius:6px; border:1px solid var(--border-subtle);">
                <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:0.75rem;">
                  <span style="font-weight:700; font-size:0.80rem; color:var(--text-bright); display:flex; align-items:center; gap:0.4rem;">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
                    Web Browsing Emulation
                  </span>
                  <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; cursor:pointer;">
                    <input type="checkbox" id="psWebEnabled" checked> Enabled
                  </label>
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.65rem; margin-bottom:0.65rem;">
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Req/min (Min / Max)</label>
                    <div style="display:flex; gap:0.3rem;">
                      <input type="number" id="psWebReqMin" value="3" style="width:50%;">
                      <input type="number" id="psWebReqMax" value="10" style="width:50%;">
                    </div>
                  </div>
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Dwell sec (Min / Max)</label>
                    <div style="display:flex; gap:0.3rem;">
                      <input type="number" id="psWebDwellMin" value="5" style="width:50%;">
                      <input type="number" id="psWebDwellMax" value="25" style="width:50%;">
                    </div>
                  </div>
                </div>
                <div style="margin-bottom:0.65rem;">
                  <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; cursor:pointer;">
                    <input type="checkbox" id="psWebFetchAssets" checked> Fetch Linked Assets (CSS, JS, Images)
                  </label>
                </div>
                <div style="margin-bottom:0.65rem;">
                  <label style="font-size:0.68rem; color:var(--text-muted); display:block; margin-bottom:2px;">Specific Persona URLs (one per line)</label>
                  <textarea id="psWebUrls" class="form-input-mono" rows="3" style="width:100%; font-size:0.72rem;" placeholder="https://intranet.corp.local&#10;https://portal.range.local"></textarea>
                </div>
                <div>
                  <label style="font-size:0.68rem; color:var(--text-muted); display:block; margin-bottom:2px;">Search Keywords (comma separated)</label>
                  <input type="text" id="psWebKeywords" style="width:100%; font-size:0.72rem;" placeholder="quarterly report, benefits portal, project plan">
                </div>
              </div>

              <!-- Module B: SMB & File Share Simulation -->
              <div style="background:var(--bg-subtle); padding:1rem; border-radius:6px; border:1px solid var(--border-subtle);">
                <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:0.75rem;">
                  <span style="font-weight:700; font-size:0.80rem; color:var(--text-bright); display:flex; align-items:center; gap:0.4rem;">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
                    SMB &amp; File Share Simulation
                  </span>
                  <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; cursor:pointer;">
                    <input type="checkbox" id="psSmbEnabled" checked> Enabled
                  </label>
                </div>
                <div style="margin-bottom:0.65rem;">
                  <label style="font-size:0.68rem; color:var(--text-muted); display:block; margin-bottom:2px;">Target Share UNC / Path</label>
                  <input type="text" id="psSmbPath" class="form-input-mono" style="width:100%; font-size:0.72rem;" placeholder="\\fileserver.range.local\corporate">
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.65rem; margin-bottom:0.65rem;">
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Interval (seconds)</label>
                    <input type="number" id="psSmbInterval" value="20" style="width:100%;">
                  </div>
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Read / Write Ratio</label>
                    <div style="display:flex; gap:0.3rem;">
                      <input type="number" step="0.05" min="0" max="1" id="psSmbReadRatio" value="0.80" style="width:50%;" title="Read Ratio">
                      <input type="number" step="0.05" min="0" max="1" id="psSmbWriteRatio" value="0.20" style="width:50%;" title="Write Ratio">
                    </div>
                  </div>
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.65rem;">
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Simulated Domain</label>
                    <input type="text" id="psSmbDomain" value="CORP" style="width:100%; font-size:0.72rem;">
                  </div>
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Simulated User</label>
                    <input type="text" id="psSmbUser" value="simulated_user" style="width:100%; font-size:0.72rem;">
                  </div>
                </div>
              </div>

              <!-- Module C: Host Benign Activity & Safe Execution -->
              <div style="background:var(--bg-subtle); padding:1rem; border-radius:6px; border:1px solid var(--border-subtle);">
                <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:0.75rem;">
                  <span style="font-weight:700; font-size:0.80rem; color:var(--text-bright); display:flex; align-items:center; gap:0.4rem;">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2"><polyline points="4 17 10 11 4 5"></polyline><line x1="12" y1="19" x2="20" y2="19"></line></svg>
                    Host Benign Activity &amp; Tools
                  </span>
                  <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; cursor:pointer;">
                    <input type="checkbox" id="psHostEnabled" checked> Enabled
                  </label>
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:0.65rem; margin-bottom:0.65rem;">
                  <div>
                    <label style="font-size:0.68rem; color:var(--text-muted);">Execution Interval (sec)</label>
                    <input type="number" id="psHostInterval" value="15" style="width:100%;">
                  </div>
                  <div style="display:flex; flex-direction:column; justify-content:center; gap:0.3rem;">
                    <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.70rem; cursor:pointer;">
                      <input type="checkbox" id="psHostDocs" checked> Simulate Office Docs
                    </label>
                    <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.70rem; cursor:pointer;" title="Strictly creates files in Documents, Downloads & Desktop with zero tool footprints">
                      <input type="checkbox" id="psHostTemp" checked> User Files (Documents, Downloads, Desktop)
                    </label>
                  </div>
                </div>
                <div>
                  <label style="font-size:0.68rem; color:var(--text-muted); display:block; margin-bottom:2px;">Safe Benign Processes (comma separated)</label>
                  <textarea id="psHostProcs" class="form-input-mono" rows="3" style="width:100%; font-size:0.72rem;" placeholder="whoami, hostname, netstat -ano, ipconfig, systeminfo"></textarea>
                </div>
              </div>

              <!-- Module D: ICMP / Ping Keepalives -->
              <div style="background:var(--bg-subtle); padding:1rem; border-radius:6px; border:1px solid var(--border-subtle);">
                <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:0.75rem;">
                  <span style="font-weight:700; font-size:0.80rem; color:var(--text-bright); display:flex; align-items:center; gap:0.4rem;">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#a78bfa" stroke-width="2"><circle cx="12" cy="12" r="1"></circle><path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path></svg>
                    ICMP / Ping Keepalives
                  </span>
                  <label style="display:flex; align-items:center; gap:0.4rem; font-size:0.72rem; cursor:pointer;">
                    <input type="checkbox" id="psPingEnabled" checked> Enabled
                  </label>
                </div>
                <div style="margin-bottom:0.65rem;">
                  <label style="font-size:0.68rem; color:var(--text-muted);">Ping Interval (seconds)</label>
                  <input type="number" id="psPingInterval" value="15" style="width:100%;">
                </div>
                <div>
                  <label style="font-size:0.68rem; color:var(--text-muted); display:block; margin-bottom:2px;">Target IP / Hostnames (comma separated)</label>
                  <input type="text" id="psPingTargets" class="form-input-mono" style="width:100%; font-size:0.72rem;" placeholder="10.0.0.1, 10.0.1.10, gateway.local">
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. Automated Emulation Schedule Windows (Direct UI for schedules.json) -->
        <div class="card">
          <div class="card-header">
            <div>
              <span class="card-title">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#38bdf8" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                Automated Emulation Schedules (schedules.json)
              </span>
              <div style="font-size:0.70rem; color:var(--text-muted); margin-top:2px;">Configure scheduled shifts in user emulation intensity and persona profiles directly through the dashboard.</div>
            </div>
            <div style="display:flex; gap:0.5rem;">
              <button class="btn btn-secondary" onclick="addNewScheduleRow()">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                Add Schedule
              </button>
              <button class="btn btn-primary" onclick="saveEmulationSchedules()">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                Save Schedules
              </button>
            </div>
          </div>
          <div class="card-body" style="padding:0;">
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Schedule Window Name</th>
                    <th>Time Window (UTC)</th>
                    <th>Assigned Persona</th>
                    <th>Intensity Multiplier</th>
                    <th>Status</th>
                    <th style="text-align:right;">Actions</th>
                  </tr>
                </thead>
                <tbody id="schedulesTableBody">
                  <tr>
                    <td colspan="6" style="text-align:center; padding:1.5rem; color:var(--text-subtle);">Loading emulation schedules...</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <!-- VIEW 3: WEB CORPUS & TARGETS -->
      <section class="view-panel" id="view-corpus">
        <div class="card">
          <div class="card-header">
            <span class="card-title">Web Browsing Corpus (web_corpus.txt)</span>
            <button class="btn btn-primary" onclick="saveWebCorpus()">Save &amp; Broadcast Corpus</button>
          </div>
          <div class="card-body">
            <textarea class="editor-textarea" id="corpusEditor" spellcheck="false" placeholder="https://portal.range.local
https://wiki.corp.local
https://jira.corp.local
https://en.wikipedia.org"></textarea>
            <span class="form-help">Strictly one site per line. Must start with http:// or https://. Simple text list without JSON.</span>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <span class="card-title">Scenario Wordlist (wordlist.txt)</span>
            <button class="btn btn-primary" onclick="saveWordlist()">Save Wordlist</button>
          </div>
          <div class="card-body">
            <textarea class="editor-textarea" id="wordlistEditor" spellcheck="false" placeholder="executive
quarterly
briefing
credentials
budget
confidential"></textarea>
            <span class="form-help">Strictly one word per line. Agents generate realistic local files and search queries based on this wordlist.</span>
          </div>
        </div>
      </section>

      <!-- VIEW 4: SIMPLIFIED NETWORK TOPOLOGY -->
      <section class="view-panel" id="view-topology">
        <!-- LIVE INTERACTIVE TOPOLOGY GRAPH -->
        <div class="card">
          <div class="card-header">
            <div style="display:flex; flex-direction:column; gap:0.15rem;">
              <span class="card-title">LIVE NETWORK TOPOLOGY MAP</span>
              <span style="font-size:0.70rem; color:var(--text-muted);">Real-time topology graph rendering uplink gateways and connected host agents.</span>
            </div>
            <span class="status-badge status-running" style="font-family:'JetBrains Mono', monospace;">AUTONOMOUS ROUTING</span>
          </div>
          <div class="card-body" style="padding:1rem; overflow-x:auto;">
            <svg id="topologySvgMap" viewBox="0 0 900 240" style="width:100%; min-width:680px; height:auto; background:var(--bg-void); border-radius:6px; border:1px solid var(--border-subtle);">
              <!-- Rendered via renderTopologyMap() -->
            </svg>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <div style="display:flex; flex-direction:column; gap:0.15rem;">
              <span class="card-title">NETWORK TOPOLOGY GATEWAYS</span>
              <span style="font-size:0.70rem; color:var(--text-muted);">Simplified single-network routing nodes. No complex JSON schemas required.</span>
            </div>
            <div style="display:flex; gap:0.5rem;">
              <button class="btn btn-secondary" style="padding:0.35rem 0.65rem; font-size:0.72rem;" onclick="addTopologyNode()">+ Add Gateway Node</button>
              <button class="btn btn-primary" onclick="saveTopologyFromTable()">Save Topology</button>
            </div>
          </div>
          <div class="table-wrap">
            <table class="data-table" id="topologyTable">
              <thead>
                <tr>
                  <th style="width:25%;">Node Name</th>
                  <th style="width:25%;">Control IP</th>
                  <th style="width:25%;">Subnet Block</th>
                  <th style="width:15%;">Role / Description</th>
                  <th style="width:10%; text-align:right;">Action</th>
                </tr>
              </thead>
              <tbody id="topologyTableBody">
                <!-- Dynamically populated -->
              </tbody>
            </table>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <span class="card-title">ADVANCED TOPOLOGY JSON (OPTIONAL)</span>
            <button class="btn btn-secondary" style="padding:0.35rem 0.65rem; font-size:0.72rem;" onclick="toggleRawTopology()">Toggle JSON Editor</button>
          </div>
          <div class="card-body" id="rawTopologyWrap" style="display:none;">
            <textarea class="editor-textarea" id="topologyEditor" spellcheck="false"></textarea>
            <div style="margin-top:0.75rem; display:flex; justify-content:flex-end;">
              <button class="btn btn-primary" onclick="saveRawTopology()">Save Raw Topology JSON</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <div style="display:flex; align-items:center; gap:0.6rem;">
              <span class="card-title">EMULATION SCHEDULE CONFIGURATION</span>
              <span style="font-size:0.62rem; font-weight:700; color:var(--text-subtle); background:var(--bg-subtle); border:1px solid var(--border-subtle); padding:2px 6px; border-radius:4px; letter-spacing:0.04em;">OPTIONAL</span>
            </div>
            <button class="btn btn-secondary" onclick="saveSchedules()">Save Optional Schedules</button>
          </div>
          <div class="card-body">
            <textarea class="editor-textarea" id="schedulesEditor" spellcheck="false" placeholder="[ (Optional) Leave empty if custom time-based shifts are not needed ]"></textarea>
            <span class="form-help">Optional schedule definition for time-based operational shifts. User Emulation runs autonomously without requiring schedules to be configured.</span>
          </div>
        </div>
      </section>

      <!-- VIEW 5: ACTIVITY & AUDIT LOGS -->
      <section class="view-panel" id="view-activity">
        <div class="card">
          <div class="card-header" style="flex-wrap:wrap; gap:0.65rem;">
            <div style="display:flex; align-items:center; gap:0.75rem;">
              <span class="card-title">COMPLETE ACTIVITY &amp; AUDIT TRAIL</span>
              <span class="activity-live-badge" title="Live Auto-Polling Active"><span class="activity-live-dot"></span> LIVE</span>
            </div>
            <div style="display:flex; align-items:center; gap:0.5rem; flex-wrap:wrap;">
              <div class="filter-pills-group" id="activityFilterPills">
                <button class="filter-pill active" onclick="setAuditFilter('ALL', this)">ALL</button>
                <button class="filter-pill" onclick="setAuditFilter('HTTP', this)">HTTP</button>
                <button class="filter-pill" onclick="setAuditFilter('SMB', this)">SMB</button>
                <button class="filter-pill" onclick="setAuditFilter('FILE', this)">FILE</button>
                <button class="filter-pill" onclick="setAuditFilter('ICMP', this)">ICMP</button>
                <button class="filter-pill" onclick="setAuditFilter('CMD', this)">CMD</button>
              </div>
              <input type="text" class="activity-search-input" id="auditSearchInput" placeholder="Filter audit logs..." oninput="handleAuditSearch(this.value)">
              <button class="btn btn-secondary" style="padding:0.30rem 0.60rem; font-size:0.70rem;" onclick="loadFullAuditLog()">Refresh</button>
              <button class="btn btn-primary" style="padding:0.30rem 0.60rem; font-size:0.70rem;" onclick="exportAuditJson()">Export JSON</button>
            </div>
          </div>
          <div class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Timestamp (UTC)</th>
                  <th>Origin Host</th>
                  <th>Protocol</th>
                  <th>Action Summary</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody id="fullAuditBody">
                <tr>
                  <td colspan="5" style="text-align:center; color:var(--text-muted); padding:2rem;">Awaiting audit logs...</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </main>
  </div>

  <!-- ADMIN CREDENTIALS MODAL -->
  <div class="modal-overlay" id="adminModal">
    <div class="modal-card">
      <div class="card-header">
        <span class="card-title">Administrative Credentials</span>
        <button class="btn btn-secondary" style="padding:0.25rem 0.5rem; font-size:0.70rem;" onclick="closeModal('adminModal')">&times;</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">Current Password</label>
          <input type="password" class="form-input" id="oldAdminPass" placeholder="Current password">
        </div>
        <div class="form-group">
          <label class="form-label">New Username (Optional)</label>
          <input type="text" class="form-input" id="newAdminUser" placeholder="Leave blank to keep current">
        </div>
        <div class="form-group">
          <label class="form-label">New Password</label>
          <input type="password" class="form-input" id="newAdminPass" placeholder="Minimum 8 characters">
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="closeModal('adminModal')">Cancel</button>
        <button class="btn btn-primary" onclick="submitAdminPassword()">Update Credentials</button>
      </div>
    </div>
  </div>

  <!-- AGENT TELEMETRY INSPECTOR MODAL -->
  <div class="modal-overlay" id="agentInspectModal">
    <div class="modal-card" style="width:740px; max-width:92vw;">
      <div class="card-header">
        <div style="display:flex; align-items:center; gap:0.65rem;">
          <span class="card-title" id="inspectModalTitle">Host Endpoint Inspection</span>
          <span class="os-badge" id="inspectModalOsBadge">Windows</span>
        </div>
        <button class="btn btn-secondary" style="padding:0.25rem 0.5rem; font-size:0.70rem;" onclick="closeModal('agentInspectModal')">&times;</button>
      </div>
      <div class="modal-body" id="inspectModalBody" style="max-height:75vh; overflow-y:auto; gap:1.10rem;">
        <!-- Dynamically rendered -->
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="closeModal('agentInspectModal')">Close</button>
        <button class="btn btn-primary" id="inspectModalShellBtn" onclick="shellFromInspect()">Open Stealth Shell</button>
      </div>
    </div>
  </div>

  <!-- EXECUTIVE REPORT MODAL -->
  <div class="modal-overlay" id="executiveReportModal">
    <div class="modal-card" style="width:700px; max-width:92vw;">
      <div class="card-header">
        <span class="card-title">Cyber Range Exercise Executive Summary Brief</span>
        <button class="btn btn-secondary" style="padding:0.25rem 0.5rem; font-size:0.70rem;" onclick="closeModal('executiveReportModal')">&times;</button>
      </div>
      <div class="modal-body">
        <div style="font-size:0.72rem; color:var(--text-muted); margin-bottom:0.4rem;">
          Compiled executive brief capturing range telemetry, active personas, host inventory, and confirmed zero-footprint artifact cleanup.
        </div>
        <textarea class="editor-textarea" id="executiveReportContent" style="min-height:300px; font-size:0.74rem;" readonly></textarea>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick="closeModal('executiveReportModal')">Close</button>
        <button class="btn btn-primary" onclick="downloadExecutiveReport()">Download Brief (.md)</button>
      </div>
    </div>
  </div>

  <!-- KEYBOARD SHORTCUTS MODAL -->
  <div class="modal-overlay" id="shortcutsModal">
    <div class="modal-card" style="width:480px; max-width:90vw;">
      <div class="card-header">
        <span class="card-title">RangeForge Keyboard Shortcuts</span>
        <button class="btn btn-secondary" style="padding:0.25rem 0.5rem; font-size:0.70rem;" onclick="closeModal('shortcutsModal')">&times;</button>
      </div>
      <div class="modal-body">
        <div class="shortcuts-list" style="display:flex; flex-direction:column; gap:0.65rem;">
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Toggle Navigation Sidebar</span><kbd class="shortcut-kbd">Ctrl + B</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Toggle Theme (Sapphire / Carbon)</span><kbd class="shortcut-kbd">T</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Overview Console</span><kbd class="shortcut-kbd">1</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Connected Fleet &amp; Personas</span><kbd class="shortcut-kbd">2</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Emulation Engine Dynamics</span><kbd class="shortcut-kbd">3</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Web Corpus &amp; Targets</span><kbd class="shortcut-kbd">4</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Network Topology</span><kbd class="shortcut-kbd">5</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Activity &amp; Audit Trail</span><kbd class="shortcut-kbd">6</kbd></div>
          <div style="display:flex; justify-content:space-between; align-items:center;"><span style="color:var(--text-bright);">Dismiss Modal / Drawers</span><kbd class="shortcut-kbd">Escape</kbd></div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-primary" onclick="closeModal('shortcutsModal')">Got It</button>
      </div>
    </div>
  </div>

  <!-- TOAST POPUP -->
  <div class="toast-popup" id="toastPopup">Notification message</div>

  <script>
    // =========================================================================
    // RANGEFORGE ENTERPRISE CLIENT ENGINE (NAVY BLUE OPERATIONS)
    // =========================================================================

    let currentPath = '/';
    let rangeData = null;
    let fleetAgentsCache = [];
    let isRequestBusy = false;

    // Rolling telemetry histories for live SVG sparklines
    let cpuHistory = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
    let ramHistory = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];

    // Session ticker state
    let sessionStartTime = null;
    let sessionRunning = false;
    let cumulativeActions = 0;

    // Toast notification
    function showToast(msg) {
      const t = document.getElementById('toastPopup');
      if (!t) return;
      t.innerText = msg;
      t.classList.add('show');
      setTimeout(() => { t.classList.remove('show'); }, 3200);
    }

    // Modal helpers
    function openAdminModal() {
      const m = document.getElementById('adminModal');
      if (m) m.classList.add('active');
    }
    function openShortcutsModal() {
      const m = document.getElementById('shortcutsModal');
      if (m) m.classList.add('active');
    }
    function closeModal(id) {
      const m = document.getElementById(id);
      if (m) m.classList.remove('active');
    }

    // Live SVG Sparklines Engine
    function updateSparklines(cpuVal, ramVal) {
      const c = Math.min(100, Math.max(0, Math.round(cpuVal)));
      const r = Math.min(100, Math.max(0, Math.round(ramVal)));
      cpuHistory.push(c);
      if (cpuHistory.length > 15) cpuHistory.shift();
      ramHistory.push(r);
      if (ramHistory.length > 15) ramHistory.shift();

      renderSparkline('sparklineCpuPoly', 'sparklineCpuLine', cpuHistory);
      renderSparkline('sparklineRamPoly', 'sparklineRamLine', ramHistory);
    }

    function renderSparkline(polyId, lineId, data) {
      const poly = document.getElementById(polyId);
      const line = document.getElementById(lineId);
      if (!poly || !line || data.length < 2) return;
      const width = 120;
      const height = 28;
      const step = width / (data.length - 1);
      let points = [];
      for (let i = 0; i < data.length; i++) {
        const x = Math.round(i * step);
        const y = Math.round(height - ((data[i] / 100) * (height - 4)) - 2);
        points.push(x + ',' + y);
      }
      line.setAttribute('points', points.join(' '));
      poly.setAttribute('points', '0,' + height + ' ' + points.join(' ') + ' ' + width + ',' + height);
    }

    // Operations Deck Live Session Ticker
    function updateSessionTicker() {
      const timeEl = document.getElementById('deckSessionTime');
      const countEl = document.getElementById('deckActionCount');
      const dotEl = document.getElementById('deckPulseDot');
      if (!timeEl || !countEl) return;
      if (sessionRunning && sessionStartTime) {
        const elapsedSec = Math.floor((Date.now() - sessionStartTime) / 1000);
        const hrs = String(Math.floor(elapsedSec / 3600)).padStart(2, '0');
        const mins = String(Math.floor((elapsedSec % 3600) / 60)).padStart(2, '0');
        const secs = String(elapsedSec % 60).padStart(2, '0');
        timeEl.innerText = 'SESSION: ' + hrs + ':' + mins + ':' + secs;
        if (dotEl) dotEl.style.opacity = '1';
      } else {
        timeEl.innerText = 'SESSION: 00:00:00';
        if (dotEl) dotEl.style.opacity = '0.35';
      }
      countEl.innerText = 'THROUGHPUT: ' + cumulativeActions + ' OPS';
    }

    // Host Telemetry Inspector Modal
    let inspectingAgentId = null;

    function inspectAgent(agentId) {
      inspectingAgentId = agentId;
      const a = (fleetAgentsCache || []).find(x => x.id === agentId);
      if (!a) {
        showToast('Agent not found');
        return;
      }
      const titleEl = document.getElementById('inspectModalTitle');
      const badgeEl = document.getElementById('inspectModalOsBadge');
      const bodyEl = document.getElementById('inspectModalBody');
      if (titleEl) titleEl.innerText = (a.hostname || a.id) + ' Host Telemetry';
      if (badgeEl) badgeEl.innerText = a.platform || a.os || 'Windows';

      const cpu = a.telemetry && a.telemetry.cpu ? a.telemetry.cpu : { percent: 0, count: 4, model: 'Virtual CPU' };
      const mem = a.telemetry && a.telemetry.memory ? a.telemetry.memory : { percent: 0, total_gb: 8.0, used_gb: 2.1 };
      const disk = a.telemetry && a.telemetry.disk ? a.telemetry.disk : { percent: 42, total_gb: 120.0, used_gb: 50.4 };
      const net = (a.telemetry && a.telemetry.network && a.telemetry.network.interfaces) ? a.telemetry.network.interfaces : [
        { name: 'eth0', ip: a.primary_ip || '10.0.10.15', mac: '00:1A:2B:3C:4D:5E', mtu: 1500 }
      ];

      const netTools = (a.network_tools && a.network_tools.length > 0) ? a.network_tools : ['ping', 'curl', 'nslookup', 'netstat'];
      const hostTools = (a.host_tools && a.host_tools.length > 0) ? a.host_tools : ['powershell', 'cmd', 'whoami', 'hostname'];

      let html = '';
      // Row 1: System Identifiers
      html += '<div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap:0.75rem; background:var(--bg-subtle); padding:0.85rem; border-radius:6px; border:1px solid var(--border-subtle);">';
      html += '<div><span style="font-size:0.68rem; color:var(--text-muted); display:block;">HOST IDENTIFIER</span><strong class="mono" style="font-size:0.78rem;">' + escapeHtml(a.id) + '</strong></div>';
      html += '<div><span style="font-size:0.68rem; color:var(--text-muted); display:block;">PRIMARY IP</span><span class="mono" style="font-size:0.78rem; color:var(--text-bright);">' + escapeHtml(a.primary_ip || '127.0.0.1') + '</span></div>';
      html += '<div><span style="font-size:0.68rem; color:var(--text-muted); display:block;">STATUS</span><span class="status-badge ' + ((a.status || '').toLowerCase() === 'online' ? 'status-running' : 'status-stopped') + '">' + escapeHtml(a.status || 'OFFLINE') + '</span></div>';
      html += '<div><span style="font-size:0.68rem; color:var(--text-muted); display:block;">ASSIGNED PERSONA</span><span class="status-badge status-standby">' + escapeHtml(a.assigned_persona || 'office_worker') + '</span></div>';
      html += '</div>';

      // Row 2: Live Resource Utilization Meters
      html += '<div>';
      html += '<h4 style="font-size:0.74rem; text-transform:uppercase; color:var(--text-muted); margin-bottom:0.5rem; letter-spacing:0.04em;">Live Resource Utilization</h4>';
      html += '<div style="display:flex; flex-direction:column; gap:0.65rem;">';
      
      // CPU Bar
      const cpuPct = Math.round(cpu.percent || 0);
      html += '<div>';
      html += '<div style="display:flex; justify-content:space-between; font-size:0.72rem; margin-bottom:0.25rem;"><span>CPU Load (' + (cpu.count || 'Multi') + ' Cores)</span><strong class="mono" style="color:#60a5fa;">' + cpuPct + '%</strong></div>';
      html += '<div style="height:6px; background:var(--bg-subtle); border-radius:3px; overflow:hidden; border:1px solid var(--border-subtle);"><div style="height:100%; width:' + cpuPct + '%; background:#3b82f6; border-radius:3px; transition:width 0.3s ease;"></div></div>';
      html += '</div>';

      // RAM Bar
      const ramPct = Math.round(mem.percent || 0);
      const ramUsed = (mem.used_gb ? mem.used_gb.toFixed(1) : ((mem.total_gb || 8) * ramPct / 100).toFixed(1)) + ' GB';
      const ramTotal = (mem.total_gb ? mem.total_gb.toFixed(1) : '8.0') + ' GB';
      html += '<div>';
      html += '<div style="display:flex; justify-content:space-between; font-size:0.72rem; margin-bottom:0.25rem;"><span>Memory Utilization (' + ramUsed + ' / ' + ramTotal + ')</span><strong class="mono" style="color:#a78bfa;">' + ramPct + '%</strong></div>';
      html += '<div style="height:6px; background:var(--bg-subtle); border-radius:3px; overflow:hidden; border:1px solid var(--border-subtle);"><div style="height:100%; width:' + ramPct + '%; background:#8b5cf6; border-radius:3px; transition:width 0.3s ease;"></div></div>';
      html += '</div>';

      // Disk Bar
      const diskPct = Math.round(disk.percent || 0);
      const diskUsed = (disk.used_gb ? disk.used_gb.toFixed(1) : '45.0') + ' GB';
      const diskTotal = (disk.total_gb ? disk.total_gb.toFixed(1) : '120.0') + ' GB';
      html += '<div>';
      html += '<div style="display:flex; justify-content:space-between; font-size:0.72rem; margin-bottom:0.25rem;"><span>Storage Pool (' + diskUsed + ' / ' + diskTotal + ')</span><strong class="mono" style="color:#34d399;">' + diskPct + '%</strong></div>';
      html += '<div style="height:6px; background:var(--bg-subtle); border-radius:3px; overflow:hidden; border:1px solid var(--border-subtle);"><div style="height:100%; width:' + diskPct + '%; background:#10b981; border-radius:3px; transition:width 0.3s ease;"></div></div>';
      html += '</div>';

      html += '</div></div>';

      // Row 3: Network Interfaces
      html += '<div>';
      html += '<h4 style="font-size:0.74rem; text-transform:uppercase; color:var(--text-muted); margin-bottom:0.5rem; letter-spacing:0.04em;">Network Adapters</h4>';
      html += '<div class="table-wrap"><table class="data-table" style="font-size:0.72rem;">';
      html += '<thead><tr><th>Adapter</th><th>IPv4 / IPv6</th><th>Hardware MAC</th><th>MTU</th></tr></thead><tbody>';
      net.forEach(ni => {
        html += '<tr><td><strong>' + escapeHtml(ni.name || 'eth0') + '</strong></td><td class="mono">' + escapeHtml(ni.ip || a.primary_ip || '10.0.0.15') + '</td><td class="mono" style="color:var(--text-muted);">' + escapeHtml(ni.mac || '00:50:56:xx:xx:xx') + '</td><td class="mono">' + (ni.mtu || 1500) + '</td></tr>';
      });
      html += '</tbody></table></div></div>';

      // Row 4: Detected UE Emulation Tooling Matrix
      html += '<div>';
      html += '<h4 style="font-size:0.74rem; text-transform:uppercase; color:var(--text-muted); margin-bottom:0.5rem; letter-spacing:0.04em;">Available UE Tool Capabilities</h4>';
      html += '<div style="display:flex; flex-direction:column; gap:0.4rem;">';
      html += '<div style="display:flex; flex-wrap:wrap; gap:4px; align-items:center;"><span style="font-size:0.68rem; color:var(--text-muted); width:80px;">Network:</span>';
      netTools.forEach(t => { html += '<span class="tool-badge tool-badge-net">🌐 ' + escapeHtml(t) + '</span>'; });
      html += '</div>';
      html += '<div style="display:flex; flex-wrap:wrap; gap:4px; align-items:center;"><span style="font-size:0.68rem; color:var(--text-muted); width:80px;">Host/Shell:</span>';
      hostTools.forEach(t => { html += '<span class="tool-badge tool-badge-host">💻 ' + escapeHtml(t) + '</span>'; });
      html += '</div>';
      html += '</div></div>';

      if (bodyEl) bodyEl.innerHTML = html;
      const modal = document.getElementById('agentInspectModal');
      if (modal) modal.classList.add('active');
    }

    function shellFromInspect() {
      if (!inspectingAgentId) return;
      closeModal('agentInspectModal');
      quickShellHost(inspectingAgentId);
    }

    // Executive Report Generator & Exporter
    function generateExecutiveReport() {
      const now = new Date();
      const r = rangeData || { name: 'Cyber Range', primary_cidr: '10.0.0.0/16', state: 'STANDBY', intensity: 'Medium' };
      const agents = fleetAgentsCache || [];
      const onlineCount = agents.filter(a => (a.status || '').toLowerCase() === 'online').length;
      const totalCount = agents.length;
      
      let winCount = 0, linCount = 0, bsdCount = 0;
      let personasCount = {};
      agents.forEach(a => {
        const os = (a.platform || a.os || '').toLowerCase();
        if (os.includes('win')) winCount++;
        else if (os.includes('linux') || os.includes('ubuntu') || os.includes('debian') || os.includes('centos')) linCount++;
        else if (os.includes('bsd')) bsdCount++;
        const p = a.assigned_persona || 'office_worker';
        personasCount[p] = (personasCount[p] || 0) + 1;
      });

      const stFilesCreated = document.getElementById('statFilesCreated') ? document.getElementById('statFilesCreated').innerText : '0';
      const stFilesDeleted = document.getElementById('statFilesDeleted') ? document.getElementById('statFilesDeleted').innerText : '0';

      let text = '# RANGEFORGE CYBER RANGE OPERATIONS // EXECUTIVE BRIEF\n';
      text += 'Generated: ' + now.toUTCString() + '\n';
      text += 'Classification: UNCLASSIFIED // RANGE EXERCISE DIRECTIVE\n';
      text += '--------------------------------------------------------------------------------\n\n';
      text += '## 1. EXERCISE ENVIRONMENT & TELEMETRY\n';
      text += '- Active Cyber Range: ' + (r.name || 'Cyber Range') + '\n';
      text += '- Control Subnet:     ' + (r.primary_cidr || '10.0.0.0/16') + '\n';
      text += '- Operational State:  ' + (r.state || 'STANDBY').toUpperCase() + '\n';
      text += '- Traffic Intensity:  ' + (r.intensity || 'Medium (1.0x)') + '\n';
      text += '- Average Fleet CPU:  ' + (r.avg_cpu ? Math.round(r.avg_cpu) + '%' : '0%') + '\n';
      text += '- Average Fleet RAM:  ' + (r.avg_ram ? Math.round(r.avg_ram) + '%' : '0%') + '\n\n';
      text += '## 2. HOST FLEET INVENTORY (' + onlineCount + '/' + totalCount + ' ONLINE)\n';
      text += '- Windows Endpoints:  ' + winCount + '\n';
      text += '- Linux Endpoints:    ' + linCount + '\n';
      text += '- BSD/Other Gateways: ' + bsdCount + '\n\n';
      text += '### Detailed Host Roster:\n';
      if (agents.length === 0) {
        text += '*(No agents registered yet)*\n';
      } else {
        agents.forEach((a, i) => {
          text += (i + 1) + '. Host: ' + (a.hostname || a.id) + ' | IP: ' + (a.primary_ip || '127.0.0.1') + ' | OS: ' + (a.platform || a.os || 'Unknown') + ' | Persona: ' + (a.assigned_persona || 'office_worker') + ' | Status: ' + (a.status || 'offline') + '\n';
        });
      }
      text += '\n## 3. ACTIVE PERSONA DISTRIBUTION\n';
      for (const [p, count] of Object.entries(personasCount)) {
        text += '- ' + p.replace('_', ' ').toUpperCase() + ': ' + count + ' agent(s)\n';
      }
      text += '\n## 4. ARTIFACT PURGE & ZERO-FOOTPRINT VERIFICATION\n';
      text += '- Simulated Files Generated: ' + stFilesCreated + '\n';
      text += '- Artifacts Cleaned/Purged:  ' + stFilesDeleted + '\n';
      text += '- Zero Footprint Policy:     STRICT (Cleaned on Stop/Pause, no command history logged)\n\n';
      text += '--------------------------------------------------------------------------------\n';
      text += 'End of RangeForge Executive Brief.\n';

      const reportEl = document.getElementById('executiveReportContent');
      if (reportEl) reportEl.value = text;
      return text;
    }

    function openExecutiveReportModal() {
      generateExecutiveReport();
      const m = document.getElementById('executiveReportModal');
      if (m) m.classList.add('active');
    }

    function downloadExecutiveReport() {
      const text = generateExecutiveReport();
      const blob = new Blob([text], { type: 'text/markdown;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'RangeForge_Executive_Brief_' + new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19) + '.md';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      showToast('Executive brief downloaded');
    }

    // Sidebar Collapsing & Responsive Drawer
    function initSidebar() {
      const sb = document.getElementById('appSidebar');
      if (!sb) return;
      if (window.innerWidth <= 900) {
        sb.classList.add('collapsed');
      } else {
        const saved = localStorage.getItem('rangeforge_sidebar_collapsed');
        if (saved === 'true') {
          sb.classList.add('collapsed');
        }
      }
    }
    function toggleSidebar() {
      const sb = document.getElementById('appSidebar');
      const bd = document.getElementById('sidebarBackdrop');
      if (!sb) return;
      sb.classList.toggle('collapsed');
      if (window.innerWidth <= 900) {
        if (bd) bd.classList.toggle('active', !sb.classList.contains('collapsed'));
      } else {
        localStorage.setItem('rangeforge_sidebar_collapsed', sb.classList.contains('collapsed'));
      }
    }
    function closeSidebarMobile() {
      if (window.innerWidth <= 900) {
        const sb = document.getElementById('appSidebar');
        const bd = document.getElementById('sidebarBackdrop');
        if (sb) sb.classList.add('collapsed');
        if (bd) bd.classList.remove('active');
      }
    }
    window.addEventListener('resize', () => {
      const bd = document.getElementById('sidebarBackdrop');
      const sb = document.getElementById('appSidebar');
      if (window.innerWidth > 900) {
        if (bd) bd.classList.remove('active');
        const saved = localStorage.getItem('rangeforge_sidebar_collapsed');
        if (sb) {
          if (saved === 'true') sb.classList.add('collapsed');
          else sb.classList.remove('collapsed');
        }
      }
    });
    document.addEventListener('keydown', (e) => {
      const activeTag = document.activeElement ? document.activeElement.tagName.toLowerCase() : '';
      const isInput = (activeTag === 'input' || activeTag === 'textarea' || activeTag === 'select');

      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
        e.preventDefault();
        toggleSidebar();
        return;
      }
      if (e.key === 'Escape') {
        closeSidebarMobile();
        closeModal('adminModal');
        closeModal('agentInspectModal');
        closeModal('executiveReportModal');
        closeModal('shortcutsModal');
        return;
      }
      if (isInput) return;

      if (e.key === '?' || (e.shiftKey && e.key === '/')) {
        e.preventDefault();
        openShortcutsModal();
      } else if (e.key.toLowerCase() === 't') {
        e.preventDefault();
        toggleTheme();
      } else if (e.key === '1') {
        navigatePage('/');
      } else if (e.key === '2') {
        navigatePage('/fleet');
      } else if (e.key === '3') {
        navigatePage('/emulation');
      } else if (e.key === '4') {
        navigatePage('/corpus');
      } else if (e.key === '5') {
        navigatePage('/topology');
      } else if (e.key === '6') {
        navigatePage('/activity');
      }
    });

    // Page Navigation
    function navigatePage(path, ev) {
      if (ev) ev.preventDefault();
      history.pushState(null, '', path);
      applyActivePage(path);
      closeSidebarMobile();
      return false;
    }
    window.addEventListener('popstate', () => {
      applyActivePage(window.location.pathname);
    });

    function applyActivePage(path) {
      currentPath = path;
      const navLinks = document.querySelectorAll('.nav-item');
      navLinks.forEach(n => n.classList.remove('active'));

      const panels = document.querySelectorAll('.view-panel');
      panels.forEach(p => p.classList.remove('active'));

      const hTitle = document.getElementById('mainHeaderTitle');
      const hSub = document.getElementById('mainHeaderSubtitle');

      if (path === '/' || path.includes('overview')) {
        document.getElementById('nav-overview').classList.add('active');
        document.getElementById('view-overview').classList.add('active');
        if (hTitle) hTitle.innerText = 'Overview Console';
        if (hSub) hSub.innerText = 'Operational Situational Awareness // Live Metrics Deck';
        loadOverviewData();
      } else if (path.includes('fleet') || path.includes('configure')) {
        const nf = document.getElementById('nav-fleet');
        if (nf) nf.classList.add('active');
        document.getElementById('view-fleet').classList.add('active');
        if (hTitle) hTitle.innerText = 'Connected Fleet & Personas';
        if (hSub) hSub.innerText = 'Exact OS Identification // Available UE Tools // Live Persona Control // Stealth Shell';
        loadFleetAgents();
      } else if (path.includes('emulation')) {
        const ne = document.getElementById('nav-emulation');
        if (ne) ne.classList.add('active');
        const ve = document.getElementById('view-emulation');
        if (ve) ve.classList.add('active');
        if (hTitle) hTitle.innerText = 'User Emulation Configuration';
        if (hSub) hSub.innerText = 'Global Traffic Dynamics // Persona Behavior Studio // Scenario Schedules';
        loadEmulationConfig();
      } else if (path.includes('corpus')) {
        document.getElementById('nav-corpus').classList.add('active');
        document.getElementById('view-corpus').classList.add('active');
        if (hTitle) hTitle.innerText = 'Web Corpus & Targets';
        if (hSub) hSub.innerText = 'HTTP/HTTPS URLs & Scenario Wordlists for File Operations';
        loadCorpusData();
      } else if (path.includes('topology')) {
        document.getElementById('nav-topology').classList.add('active');
        document.getElementById('view-topology').classList.add('active');
        if (hTitle) hTitle.innerText = 'Network Topology';
        if (hSub) hSub.innerText = 'Routing Nodes & Emulation Schedules';
        loadTopologyData();
      } else if (path.includes('activity')) {
        document.getElementById('nav-activity').classList.add('active');
        document.getElementById('view-activity').classList.add('active');
        if (hTitle) hTitle.innerText = 'Activity & Audit Logs';
        if (hSub) hSub.innerText = 'Chronological Emulation Event Stream';
        loadFullAuditLog();
      }

      // Check for deep-link modal hashes
      const hash = window.location.hash || '';
      if (hash === '#report') {
        setTimeout(openExecutiveReportModal, 150);
      } else if (hash === '#shortcuts') {
        setTimeout(openShortcutsModal, 150);
      } else if (hash.startsWith('#inspect-')) {
        const id = hash.replace('#inspect-', '');
        setTimeout(() => inspectAgent(id), 250);
      }
    }

    // Theme toggling: Ocean Sapphire (Cobalt Blue) vs Carbon Black (Stealth Operations)
    function initTheme() {
      const saved = localStorage.getItem('rangeforge_theme') || 'ocean-sapphire';
      document.documentElement.setAttribute('data-theme', saved);
    }
    function toggleTheme() {
      const cur = document.documentElement.getAttribute('data-theme') || 'ocean-sapphire';
      const next = cur === 'ocean-sapphire' ? 'carbon-black' : 'ocean-sapphire';
      document.documentElement.setAttribute('data-theme', next);
      localStorage.setItem('rangeforge_theme', next);
      showToast('Theme: ' + (next === 'ocean-sapphire' ? 'Ocean Sapphire (Cobalt Blue)' : 'Carbon Black (Stealth Ops)'));
    }


    // =========================================================================
    // USER EMULATION CONFIGURATION (NO JSON FILE REQUIRED)
    // =========================================================================
    let cachedPersonas = {};
    let cachedSchedules = [];
    let currentNoiseLevel = 'balanced';

    async function loadNoiseControl() {
      try {
        const res = await fetch('/api/v1/controller/noise_control');
        if (!res.ok) return;
        const data = await res.json();
        currentNoiseLevel = data.noise_level || 'balanced';
        if (document.getElementById('ncToolInterval')) document.getElementById('ncToolInterval').value = data.tool_interval_sec || 60;
        if (document.getElementById('ncHeartbeatSec')) document.getElementById('ncHeartbeatSec').value = data.heartbeat_sec || 5;
        if (document.getElementById('ncPauseTools')) document.getElementById('ncPauseTools').checked = !!data.pause_tools;
        highlightActiveNoisePreset(currentNoiseLevel);
      } catch (err) {
        console.error('Error loading noise control:', err);
      }
    }

    function highlightActiveNoisePreset(level) {
      ['presetStealth', 'presetBalanced', 'presetActive'].forEach(id => {
        const el = document.getElementById(id);
        if (el) {
          el.style.borderColor = 'var(--border-subtle)';
          el.style.background = 'var(--bg-subtle)';
        }
      });
      const activeId = level === 'stealth' ? 'presetStealth' : (level === 'active' ? 'presetActive' : 'presetBalanced');
      const activeEl = document.getElementById(activeId);
      if (activeEl) {
        activeEl.style.borderColor = '#38bdf8';
        activeEl.style.background = 'rgba(56,189,248,0.06)';
      }
    }

    function setNoisePreset(level) {
      currentNoiseLevel = level;
      if (level === 'stealth') {
        if (document.getElementById('ncToolInterval')) document.getElementById('ncToolInterval').value = 180;
        if (document.getElementById('ncHeartbeatSec')) document.getElementById('ncHeartbeatSec').value = 30;
        if (document.getElementById('ncPauseTools')) document.getElementById('ncPauseTools').checked = false;
      } else if (level === 'active') {
        if (document.getElementById('ncToolInterval')) document.getElementById('ncToolInterval').value = 20;
        if (document.getElementById('ncHeartbeatSec')) document.getElementById('ncHeartbeatSec').value = 5;
        if (document.getElementById('ncPauseTools')) document.getElementById('ncPauseTools').checked = false;
      } else {
        if (document.getElementById('ncToolInterval')) document.getElementById('ncToolInterval').value = 60;
        if (document.getElementById('ncHeartbeatSec')) document.getElementById('ncHeartbeatSec').value = 10;
        if (document.getElementById('ncPauseTools')) document.getElementById('ncPauseTools').checked = false;
      }
      highlightActiveNoisePreset(level);
      showToast('Posture Preset: ' + level.toUpperCase() + ' selected. Click Save to deploy to fleet.');
    }

    async function saveNoiseControl() {
      const payload = {
        noise_level: currentNoiseLevel,
        tool_interval_sec: parseInt(document.getElementById('ncToolInterval').value, 10) || 60,
        heartbeat_sec: parseInt(document.getElementById('ncHeartbeatSec').value, 10) || 5,
        pause_tools: document.getElementById('ncPauseTools').checked
      };

      try {
        const res = await fetch('/api/v1/controller/noise_control', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          showToast('Blue Team noise & polling cadence saved and broadcast to all agents.');
          loadNoiseControl();
        } else {
          showToast('Failed to save noise settings', true);
        }
      } catch (err) {
        showToast('Network error saving noise settings', true);
      }
    }

    async function purgeAllArtifactsNow() {
      if (!confirm('Purge all synthetic user files from Documents, Downloads, and Desktop across all active agents?')) {
        return;
      }
      try {
        const res = await fetch('/api/v1/controller/purge_artifacts', { method: 'POST' });
        if (res.ok) {
          showToast('Artifact purge dispatched to fleet. Zero footprint verified.');
        } else {
          showToast('Failed to dispatch purge', true);
        }
      } catch (err) {
        showToast('Network error during purge dispatch', true);
      }
    }

    async function loadEmulationConfig() {
      await Promise.all([
        loadNoiseControl(),
        loadGlobalEmulationRanges(),
        loadAllPersonasStudio(),
        loadEmulationSchedules()
      ]);
    }

    async function loadGlobalEmulationRanges() {
      try {
        const res = await fetch('/api/v1/controller/ranges');
        if (!res.ok) return;
        const data = await res.json();
        const r = (data.ranges && data.ranges.length > 0) ? data.ranges[0] : null;
        if (!r) return;

        const selIntensity = document.getElementById('cfgIntensity');
        if (selIntensity && r.intensity) {
          for (let opt of selIntensity.options) {
            if (opt.value.toLowerCase().startsWith(r.intensity.toLowerCase().slice(0, 3))) {
              selIntensity.value = opt.value;
              break;
            }
          }
        }

        const selMode = document.getElementById('cfgOperatingMode');
        if (selMode && r.operating_mode) selMode.value = r.operating_mode;

        const selEgress = document.getElementById('cfgEgressPolicy');
        if (selEgress && r.egress_policy) selEgress.value = r.egress_policy;

        if (document.getElementById('cfgDwellMin')) document.getElementById('cfgDwellMin').value = r.dwell_time_min_sec || 5;
        if (document.getElementById('cfgDwellMax')) document.getElementById('cfgDwellMax').value = r.dwell_time_max_sec || 25;
        if (document.getElementById('cfgWebReqMin')) document.getElementById('cfgWebReqMin').value = r.web_req_min_per_min || 4;
        if (document.getElementById('cfgWebReqMax')) document.getElementById('cfgWebReqMax').value = r.web_req_max_per_min || 16;
        if (document.getElementById('cfgLatencyMs')) document.getElementById('cfgLatencyMs').value = r.latency_ms || 15;
        if (document.getElementById('cfgJitterPct')) document.getElementById('cfgJitterPct').value = r.jitter_pct || 25;
        if (document.getElementById('cfgPacketLoss')) document.getElementById('cfgPacketLoss').value = (r.packet_loss_pct !== undefined) ? r.packet_loss_pct : 0.0;
      } catch (err) {
        console.error('Error loading global emulation config:', err);
      }
    }

    async function saveGlobalEmulationConfig() {
      const payload = {
        name: rangeData ? rangeData.name : 'Cyber Range',
        primary_cidr: rangeData ? rangeData.primary_cidr : '10.0.0.0/16',
        intensity: document.getElementById('cfgIntensity').value,
        operating_mode: document.getElementById('cfgOperatingMode').value,
        egress_policy: document.getElementById('cfgEgressPolicy').value,
        dwell_time_min_sec: parseInt(document.getElementById('cfgDwellMin').value, 10) || 5,
        dwell_time_max_sec: parseInt(document.getElementById('cfgDwellMax').value, 10) || 25,
        web_req_min_per_min: parseInt(document.getElementById('cfgWebReqMin').value, 10) || 4,
        web_req_max_per_min: parseInt(document.getElementById('cfgWebReqMax').value, 10) || 16,
        latency_ms: parseInt(document.getElementById('cfgLatencyMs').value, 10) || 15,
        jitter_pct: parseInt(document.getElementById('cfgJitterPct').value, 10) || 25,
        packet_loss_pct: parseFloat(document.getElementById('cfgPacketLoss').value) || 0.0
      };

      try {
        const res = await fetch('/api/v1/controller/ranges', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          showToast('Global emulation dynamics applied & persisted.');
          loadRangeDetails();
        } else {
          showToast('Failed to save emulation dynamics', true);
        }
      } catch (err) {
        showToast('Network error saving emulation dynamics', true);
      }
    }

    async function loadAllPersonasStudio() {
      try {
        const res = await fetch('/api/v1/controller/personas');
        if (!res.ok) return;
        cachedPersonas = await res.json();
        loadSelectedPersonaStudio();
      } catch (err) {
        console.error('Error loading personas:', err);
      }
    }

    function loadSelectedPersonaStudio() {
      const sel = document.getElementById('personaStudioSelect');
      if (!sel) return;
      const name = sel.value;
      const p = cachedPersonas[name] || {
        name: name,
        web_browsing: { enabled: true, requests_per_minute_min: 3, requests_per_minute_max: 10, dwell_time_min_sec: 5, dwell_time_max_sec: 25, fetch_assets: true, target_urls: [], search_keywords: [] },
        file_share: { enabled: true, interval_sec: 20, shares: [], read_ratio: 0.8, write_ratio: 0.2 },
        host_activity: { enabled: true, interval_sec: 15, safe_processes: ['whoami', 'hostname'], simulate_office_docs: true, temp_file_operations: true },
        ping: { enabled: true, interval_sec: 15, targets: ['10.0.0.1'] }
      };

      // Module A: Web Browsing
      const wb = p.web_browsing || {};
      document.getElementById('psWebEnabled').checked = (wb.enabled !== false);
      document.getElementById('psWebReqMin').value = wb.requests_per_minute_min || 3;
      document.getElementById('psWebReqMax').value = wb.requests_per_minute_max || 10;
      document.getElementById('psWebDwellMin').value = wb.dwell_time_min_sec || 5;
      document.getElementById('psWebDwellMax').value = wb.dwell_time_max_sec || 25;
      document.getElementById('psWebFetchAssets').checked = (wb.fetch_assets !== false);
      document.getElementById('psWebUrls').value = (wb.target_urls || []).join('\n');
      document.getElementById('psWebKeywords').value = (wb.search_keywords || []).join(', ');

      // Module B: SMB
      const fs = p.file_share || {};
      document.getElementById('psSmbEnabled').checked = (fs.enabled !== false);
      document.getElementById('psSmbInterval').value = fs.interval_sec || 20;
      document.getElementById('psSmbReadRatio').value = fs.read_ratio !== undefined ? fs.read_ratio : 0.80;
      document.getElementById('psSmbWriteRatio').value = fs.write_ratio !== undefined ? fs.write_ratio : 0.20;
      const firstShare = (fs.shares && fs.shares.length > 0) ? fs.shares[0] : null;
      document.getElementById('psSmbPath').value = firstShare ? firstShare.path : '\\\\fileserver.range.local\\corporate';
      document.getElementById('psSmbDomain').value = firstShare ? firstShare.domain : 'CORP';
      document.getElementById('psSmbUser').value = firstShare ? firstShare.username : 'simulated_user';

      // Module C: Host Activity
      const ha = p.host_activity || {};
      document.getElementById('psHostEnabled').checked = (ha.enabled !== false);
      document.getElementById('psHostInterval').value = ha.interval_sec || 15;
      document.getElementById('psHostDocs').checked = (ha.simulate_office_docs !== false);
      document.getElementById('psHostTemp').checked = (ha.temp_file_operations !== false);
      document.getElementById('psHostProcs').value = (ha.safe_processes || ['whoami', 'hostname', 'netstat -ano']).join(', ');

      // Module D: Ping
      const pg = p.ping || {};
      document.getElementById('psPingEnabled').checked = (pg.enabled === true);
      document.getElementById('psPingInterval').value = pg.interval_sec || 15;
      document.getElementById('psPingTargets').value = (pg.targets || ['10.0.0.1', 'gateway.local']).join(', ');
    }

    async function savePersonaStudioConfig() {
      const sel = document.getElementById('personaStudioSelect');
      if (!sel) return;
      const name = sel.value;

      const rawUrls = document.getElementById('psWebUrls').value.split('\n').map(s => s.trim()).filter(Boolean);
      const rawKw = document.getElementById('psWebKeywords').value.split(',').map(s => s.trim()).filter(Boolean);
      const rawProcs = document.getElementById('psHostProcs').value.split(',').map(s => s.trim()).filter(Boolean);
      const rawPings = document.getElementById('psPingTargets').value.split(',').map(s => s.trim()).filter(Boolean);

      const payload = {
        name: name,
        description: cachedPersonas[name] ? cachedPersonas[name].description : 'User Emulation Persona',
        web_browsing: {
          enabled: document.getElementById('psWebEnabled').checked,
          requests_per_minute_min: parseInt(document.getElementById('psWebReqMin').value, 10) || 2,
          requests_per_minute_max: parseInt(document.getElementById('psWebReqMax').value, 10) || 8,
          dwell_time_min_sec: parseInt(document.getElementById('psWebDwellMin').value, 10) || 5,
          dwell_time_max_sec: parseInt(document.getElementById('psWebDwellMax').value, 10) || 25,
          fetch_assets: document.getElementById('psWebFetchAssets').checked,
          user_agents: [],
          target_urls: rawUrls,
          search_keywords: rawKw
        },
        file_share: {
          enabled: document.getElementById('psSmbEnabled').checked,
          interval_sec: parseInt(document.getElementById('psSmbInterval').value, 10) || 20,
          read_ratio: parseFloat(document.getElementById('psSmbReadRatio').value) || 0.8,
          write_ratio: parseFloat(document.getElementById('psSmbWriteRatio').value) || 0.2,
          shares: [
            {
              path: document.getElementById('psSmbPath').value,
              domain: document.getElementById('psSmbDomain').value,
              username: document.getElementById('psSmbUser').value,
              password: ''
            }
          ]
        },
        host_activity: {
          enabled: document.getElementById('psHostEnabled').checked,
          interval_sec: parseInt(document.getElementById('psHostInterval').value, 10) || 15,
          simulate_office_docs: document.getElementById('psHostDocs').checked,
          temp_file_operations: document.getElementById('psHostTemp').checked,
          safe_processes: rawProcs
        },
        ping: {
          enabled: document.getElementById('psPingEnabled').checked,
          interval_sec: parseInt(document.getElementById('psPingInterval').value, 10) || 15,
          targets: rawPings
        }
      };

      try {
        const res = await fetch('/api/v1/controller/persona/save', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          cachedPersonas[name] = payload;
          showToast("Persona '" + name + "' profile saved & dispatched live to agents!");
        } else {
          showToast('Failed to save persona profile', true);
        }
      } catch (err) {
        showToast('Network error saving persona profile', true);
      }
    }

    async function loadEmulationSchedules() {
      try {
        const res = await fetch('/api/v1/controller/schedules');
        if (!res.ok) return;
        cachedSchedules = await res.json();
        renderSchedulesTable();
      } catch (err) {
        console.error('Error loading schedules:', err);
      }
    }

    function renderSchedulesTable() {
      const tb = document.getElementById('schedulesTableBody');
      if (!tb) return;
      if (!cachedSchedules || cachedSchedules.length === 0) {
        tb.innerHTML = '<tr><td colspan="6" style="text-align:center; padding:1.5rem; color:var(--text-subtle);">No active schedules configured. Click "+ Add Schedule" to create one.</td></tr>';
        return;
      }

      let html = '';
      cachedSchedules.forEach((s, idx) => {
        html += '<tr>' +
          '<td><input type="text" value="' + escapeHtml(s.name || '') + '" onchange="updateScheduleItem(' + idx + ', \'name\', this.value)" style="width:100%; font-size:0.75rem;"></td>' +
          '<td><input type="text" class="form-input-mono" value="' + escapeHtml(s.time_window || '08:00 - 17:00 UTC') + '" onchange="updateScheduleItem(' + idx + ', \'time_window\', this.value)" style="width:100%; font-size:0.72rem;"></td>' +
          '<td>' +
            '<select onchange="updateScheduleItem(' + idx + ', \'profile\', this.value)" style="width:100%; font-size:0.75rem;">' +
              '<option value="office_worker"' + (s.profile === 'office_worker' ? ' selected' : '') + '>Office Worker</option>' +
              '<option value="developer"' + (s.profile === 'developer' ? ' selected' : '') + '>Developer</option>' +
              '<option value="sysadmin"' + (s.profile === 'sysadmin' ? ' selected' : '') + '>Sysadmin</option>' +
              '<option value="executive"' + (s.profile === 'executive' ? ' selected' : '') + '>Executive</option>' +
              '<option value="finance"' + (s.profile === 'finance' ? ' selected' : '') + '>Finance</option>' +
              '<option value="scada_operator"' + (s.profile === 'scada_operator' ? ' selected' : '') + '>SCADA Operator</option>' +
            '</select>' +
          '</td>' +
          '<td><input type="number" step="0.1" min="0.1" max="5.0" value="' + (s.intensity || 1.0) + '" onchange="updateScheduleItem(' + idx + ', \'intensity\', parseFloat(this.value))" style="width:70px; font-size:0.75rem;"></td>' +
          '<td>' +
            '<label style="display:flex; align-items:center; gap:0.4rem; cursor:pointer;">' +
              '<input type="checkbox" ' + (s.enabled ? 'checked' : '') + ' onchange="updateScheduleItem(' + idx + ', \'enabled\', this.checked)">' +
              '<span class="status-badge ' + (s.enabled ? 'status-running' : 'status-standby') + '" style="font-size:0.62rem;">' + (s.enabled ? 'ACTIVE' : 'DISABLED') + '</span>' +
            '</label>' +
          '</td>' +
          '<td style="text-align:right;">' +
            '<button class="btn btn-danger" onclick="deleteScheduleItem(' + idx + ')" style="padding:0.25rem 0.5rem; font-size:0.70rem;">Delete</button>' +
          '</td>' +
        '</tr>';
      });
      tb.innerHTML = html;
    }

    function updateScheduleItem(idx, key, val) {
      if (cachedSchedules[idx]) {
        cachedSchedules[idx][key] = val;
      }
    }

    function addNewScheduleRow() {
      cachedSchedules.push({
        id: 'sched-' + (cachedSchedules.length + 1).toString().padStart(2, '0'),
        name: 'New Operation Window',
        time_window: '09:00 - 17:00 UTC',
        intensity: 1.0,
        profile: 'office_worker',
        enabled: true,
        status: 'active'
      });
      renderSchedulesTable();
    }

    function deleteScheduleItem(idx) {
      cachedSchedules.splice(idx, 1);
      renderSchedulesTable();
    }

    async function saveEmulationSchedules() {
      try {
        const res = await fetch('/api/v1/controller/schedules', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(cachedSchedules)
        });
        if (res.ok) {
          showToast('Emulation schedules persisted & applied!');
          renderSchedulesTable();
        } else {
          showToast('Failed to save schedules', true);
        }
      } catch (err) {
        showToast('Network error saving schedules', true);
      }
    }

    // Clocks
    function updateClocks() {
      const now = new Date();
      const utc = document.getElementById('utc-clock');
      const loc = document.getElementById('loc-clock');
      if (utc) utc.innerText = now.toUTCString().split(' ')[4] || '--:--:--';
      if (loc) loc.innerText = now.toTimeString().split(' ')[0] || '--:--:--';
    }

    // =========================================================================
    // DATA LOADERS & API CALLS
    // =========================================================================

    async function loadRangeDetails() {
      try {
        const res = await fetch('/api/v1/controller/ranges');
        if (!res.ok) return;
        const data = await res.json();
        const r = (data.ranges && data.ranges.length > 0) ? data.ranges[0] : null;
        if (!r) return;
        rangeData = r;

        // Top bar & Sidebar Range Pod
        const topLabel = document.getElementById('topActiveRangeLabel');
        const sbName = document.getElementById('sbRangeName');
        const sbCidr = document.getElementById('sbRangeCidr');
        const sbBadge = document.getElementById('sbRangeBadge');
        const headBadge = document.getElementById('headerStatusBadge');
        const deckBadge = document.getElementById('deckRangeBadge');
        const deckName = document.getElementById('deckRangeName');
        const deckCidr = document.getElementById('deckRangeCidr');
        const deckIntensity = document.getElementById('deckIntensitySelect');

        const activeName = r.name || 'Cyber Range';
        const activeCidr = r.primary_cidr || '10.0.0.0/16';

        if (topLabel) topLabel.innerText = activeName;
        if (sbName) sbName.innerText = activeName;
        if (sbCidr) sbCidr.innerText = activeCidr;
        if (deckName) deckName.innerText = activeName;
        if (deckCidr) deckCidr.innerText = activeCidr;

        if (sbBadge) setBadgeState(sbBadge, r.state);
        if (headBadge) setBadgeState(headBadge, r.state);
        if (deckBadge) setBadgeState(deckBadge, r.state);

        if (deckIntensity && r.intensity) {
          for (let opt of deckIntensity.options) {
            if (opt.value.toLowerCase().startsWith(r.intensity.toLowerCase().slice(0, 3))) {
              deckIntensity.value = opt.value;
              break;
            }
          }
        }

        // Stats Row
        const stState = document.getElementById('statOpState');
        const stEndpoints = document.getElementById('statEndpoints');
        const stIntensity = document.getElementById('statIntensity');
        const stCpu = document.getElementById('statCpu');
        const stRam = document.getElementById('statRam');

        if (stState) {
          stState.innerText = (r.state || 'STANDBY').toUpperCase();
          stState.style.color = (r.state === 'running') ? '#10b981' : ((r.state === 'paused') ? '#38bdf8' : '#f59e0b');
        }
        if (stEndpoints) {
          stEndpoints.innerText = (r.endpoints_online || 0) + ' / ' + (r.endpoints_total || 1);
          stEndpoints.style.color = '#38bdf8';
        }
        if (stIntensity && r.intensity) {
          stIntensity.innerText = r.intensity.split(' ')[0];
          stIntensity.style.color = '#f59e0b';
        }
        if (stCpu) {
          stCpu.innerText = Math.round(r.avg_cpu || 0) + '%';
          stCpu.style.color = '#60a5fa';
        }
        if (stRam) {
          stRam.innerText = Math.round(r.avg_ram || 0) + '%';
          stRam.style.color = '#a78bfa';
        }

        // Live SVG Sparklines
        updateSparklines(r.avg_cpu || 0, r.avg_ram || 0);

        // Synchronize Operations Deck Session Ticker
        const curRangeState = (r.state || '').toLowerCase();
        if (curRangeState === 'running') {
          sessionRunning = true;
          if (!sessionStartTime) sessionStartTime = Date.now();
        } else if (curRangeState === 'stopped' || curRangeState === 'emergency_stop') {
          sessionRunning = false;
          sessionStartTime = null;
        } else if (curRangeState === 'paused') {
          sessionRunning = false;
        }
        updateSessionTicker();
      } catch (e) {
        console.error('Failed to load range details:', e);
      }
    }

    async function loadOverviewData() {
      await Promise.all([
        loadRangeDetails(),
        loadFleetAgents(),
        loadRecentEvents()
      ]);
    }

    function setBadgeState(el, state) {
      const s = (state || 'stopped').toLowerCase();
      el.className = 'status-badge';
      if (s === 'running') {
        el.classList.add('status-running');
        el.innerText = 'RUNNING';
      } else if (s === 'paused') {
        el.classList.add('status-paused');
        el.innerText = 'PAUSED';
      } else if (s === 'emergency_stop' || s === 'stopped') {
        el.classList.add('status-stopped');
        el.innerText = s === 'emergency_stop' ? 'E-STOPPED' : 'STOPPED';
      } else {
        el.classList.add('status-standby');
        el.innerText = 'STANDBY';
      }
    }

    // Range state controls (START, PAUSE, STOP)
    async function setRangeState(state) {
      const stState = document.getElementById('statOpState');
      const sbBadge = document.getElementById('sbRangeBadge');
      const headBadge = document.getElementById('headerStatusBadge');
      const deckBadge = document.getElementById('deckRangeBadge');

      if (stState) {
        stState.innerText = state.toUpperCase();
        stState.style.color = (state === 'running') ? '#10b981' : ((state === 'paused') ? '#38bdf8' : '#f59e0b');
      }
      if (sbBadge) setBadgeState(sbBadge, state);
      if (headBadge) setBadgeState(headBadge, state);
      if (deckBadge) setBadgeState(deckBadge, state);

      const reqState = (state || '').toLowerCase();
      if (reqState === 'running') {
        sessionRunning = true;
        if (!sessionStartTime) sessionStartTime = Date.now();
      } else if (reqState === 'stopped' || reqState === 'emergency_stop') {
        sessionRunning = false;
        sessionStartTime = null;
      } else if (reqState === 'paused') {
        sessionRunning = false;
      }
      updateSessionTicker();

      showToast('Emulation state: ' + state.toUpperCase());
      try {
        const payload = {
          state: state,
          range_name: rangeData ? rangeData.name : 'Cyber Range'
        };
        const res = await fetch('/api/v1/controller/ranges/state', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          showToast('Operational state updated: ' + state.toUpperCase());
          loadRangeDetails();
        } else {
          showToast('Failed to update emulation state');
        }
      } catch (e) {
        showToast('Network error: ' + e);
      }
    }

    // Intensity calibration
    async function setRangeIntensity(intensity) {
      showToast('Setting intensity to ' + intensity);
      try {
        const res = await fetch('/api/v1/controller/intensity', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ intensity: intensity })
        });
        if (res.ok) {
          showToast('Intensity updated: ' + intensity);
          loadRangeDetails();
        }
      } catch (e) {
        showToast('Error updating intensity: ' + e);
      }
    }

    // =========================================================================
    // FLEET AGENTS, EXACT OS & PERSONA MANAGEMENT
    // =========================================================================

    async function loadFleetAgents() {
      try {
        const res = await fetch('/api/v1/controller/agents');
        if (!res.ok) return;
        const data = await res.json();
        const agents = data.agents || [];
        fleetAgentsCache = agents;

        const navCount = document.getElementById('navFleetCount');
        if (navCount) navCount.innerText = agents.length;

        // Render both tables: Overview summary and Fleet management
        renderOverviewFleetTable(agents);
        renderFleetDetailTable(agents);
        populateCmdTargets(agents);
        renderTopologyMap();

        // Update wordlist file metrics
        let totalCreated = 0;
        let totalDeleted = 0;
        agents.forEach(a => {
          if (a.telemetry && a.telemetry.emulation_metrics) {
            totalCreated += a.telemetry.emulation_metrics.files_created || 0;
            totalDeleted += a.telemetry.emulation_metrics.files_deleted || 0;
          }
        });
        const fcEl = document.getElementById('statFilesCreated');
        if (fcEl) {
          fcEl.innerText = totalCreated;
          fcEl.style.color = '#34d399';
        }
        const fdEl = document.getElementById('statFilesDeleted');
        if (fdEl) fdEl.innerText = totalDeleted;
      } catch (e) {
        console.error('Failed to load fleet agents:', e);
      }
    }

    function renderOverviewFleetTable(agents) {
      const tbody = document.getElementById('overviewFleetTableBody');
      if (!tbody) return;

      if (agents.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" style="text-align:center; color:var(--text-muted); padding:2rem;">No agent hosts registered on this range yet.</td></tr>';
        return;
      }

      let html = '';
      agents.forEach(a => {
        const isOnline = (a.status || '').toLowerCase() === 'online';
        const badgeClass = isOnline ? 'status-running' : 'status-stopped';
        const cpu = a.telemetry && a.telemetry.cpu ? Math.round(a.telemetry.cpu.percent) + '%' : '0%';
        const ram = a.telemetry && a.telemetry.memory ? Math.round(a.telemetry.memory.percent) + '%' : '0%';
        const exactOS = a.platform || a.os || 'Windows 11 (amd64)';

        html += '<tr>';
        html += '<td><strong>' + escapeHtml(a.hostname || a.id) + '</strong><br><span style="font-size:0.68rem; color:var(--text-muted); font-family:monospace;">' + escapeHtml(a.id) + '</span></td>';
        html += '<td><span class="os-badge">' + getOsIcon(exactOS) + ' ' + escapeHtml(exactOS) + '</span></td>';
        html += '<td><span class="copy-chip mono" data-ip="' + escapeHtml(a.primary_ip) + '" onclick="copyIp(this.dataset.ip)">' + escapeHtml(a.primary_ip || '127.0.0.1') + ' 📋</span></td>';
        html += '<td><span class="status-badge status-standby">' + escapeHtml(a.assigned_persona || 'office_worker') + '</span></td>';
        html += '<td class="mono" style="color:#60a5fa;">' + cpu + '</td>';
        html += '<td class="mono" style="color:#a78bfa;">' + ram + '</td>';
        html += '<td><span class="status-badge ' + badgeClass + '">' + escapeHtml(a.status || 'offline') + '</span></td>';
        html += '<td style="text-align:right; white-space:nowrap;">';
        html += '<button class="btn btn-secondary" style="padding:0.25rem 0.50rem; font-size:0.70rem; margin-right:4px;" data-id="' + escapeHtml(a.id) + '" onclick="inspectAgent(this.dataset.id)">Inspect</button>';
        html += '<button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.72rem;" data-id="' + escapeHtml(a.id) + '" onclick="quickShellHost(this.dataset.id)">Command</button>';
        html += '</td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    }

    function renderFleetDetailTable(agents) {
      const tbody = document.getElementById('fleetDetailTableBody');
      if (!tbody) return;

      if (agents.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center; color:var(--text-muted); padding:2rem;">No agent hosts registered on this range yet.</td></tr>';
        return;
      }

      let html = '';
      agents.forEach(a => {
        const isOnline = (a.status || '').toLowerCase() === 'online';
        const badgeClass = isOnline ? 'status-running' : 'status-stopped';
        const exactOS = a.platform || a.os || 'Windows 11 (amd64)';
        const curPersona = a.assigned_persona || 'office_worker';

        // Badges for Network and Host Tools
        let netBadges = '';
        const netTools = (a.network_tools && a.network_tools.length > 0) ? a.network_tools : ['ping', 'curl', 'nslookup', 'netstat'];
        netTools.forEach(t => {
          netBadges += '<span class="tool-badge tool-badge-net" title="Network UE Tool">🌐 ' + escapeHtml(t) + '</span>';
        });

        let hostBadges = '';
        const hostTools = (a.host_tools && a.host_tools.length > 0) ? a.host_tools : ['powershell', 'cmd', 'whoami', 'hostname'];
        hostTools.forEach(t => {
          hostBadges += '<span class="tool-badge tool-badge-host" title="Host UE Tool">💻 ' + escapeHtml(t) + '</span>';
        });

        // Persona select options
        const personas = [
          { val: 'office_worker', label: 'Office Worker' },
          { val: 'developer', label: 'Developer' },
          { val: 'sysadmin', label: 'System Admin' },
          { val: 'finance', label: 'Finance Specialist' },
          { val: 'executive', label: 'Corporate Executive' },
          { val: 'hr_specialist', label: 'HR Specialist' },
          { val: 'scada_operator', label: 'SCADA / ICS Operator' }
        ];
        let personaOptions = '';
        personas.forEach(p => {
          const sel = (p.val === curPersona) ? ' selected' : '';
          personaOptions += '<option value="' + p.val + '"' + sel + '>' + p.label + '</option>';
        });

        html += '<tr>';
        html += '<td><strong>' + escapeHtml(a.hostname || a.id) + '</strong><br><span style="font-size:0.68rem; color:var(--text-muted); font-family:monospace;">' + escapeHtml(a.id) + '</span></td>';
        html += '<td><span class="os-badge">' + getOsIcon(exactOS) + ' ' + escapeHtml(exactOS) + '</span></td>';
        html += '<td><span class="copy-chip mono" data-ip="' + escapeHtml(a.primary_ip) + '" onclick="copyIp(this.dataset.ip)">' + escapeHtml(a.primary_ip || '127.0.0.1') + ' 📋</span></td>';
        html += '<td><div style="display:flex; flex-direction:column; gap:3px;"><div>' + netBadges + '</div><div>' + hostBadges + '</div></div></td>';
        html += '<td>';
        html += '<select class="form-input form-input-mono" style="padding:0.30rem 0.50rem; font-size:0.75rem; width:100%;" data-id="' + escapeHtml(a.id) + '" onchange="changeAgentPersona(this.dataset.id, this.value)">' + personaOptions + '</select>';
        html += '</td>';
        html += '<td><span class="status-badge ' + badgeClass + '">' + escapeHtml(a.status || 'offline') + '</span></td>';
        html += '<td style="text-align:right; white-space:nowrap;">';
        html += '<button class="btn btn-secondary" style="padding:0.25rem 0.50rem; font-size:0.70rem; margin-right:4px;" data-id="' + escapeHtml(a.id) + '" onclick="inspectAgent(this.dataset.id)">Inspect</button>';
        html += '<button class="btn btn-secondary" style="padding:0.25rem 0.55rem; font-size:0.72rem;" data-id="' + escapeHtml(a.id) + '" onclick="quickShellHost(this.dataset.id)">Command</button>';
        html += '</td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    }

    function getOsIcon(osStr) {
      const s = (osStr || '').toLowerCase();
      if (s.includes('win')) return '🪟';
      if (s.includes('ubuntu') || s.includes('linux') || s.includes('debian') || s.includes('centos')) return '🐧';
      if (s.includes('freebsd') || s.includes('pfsense') || s.includes('bsd')) return '🔴';
      if (s.includes('darwin') || s.includes('mac')) return '🍎';
      return '💻';
    }

    function populateCmdTargets(agents) {
      const targetSel = document.getElementById('fleetCmdTarget');
      if (!targetSel) return;
      const curVal = targetSel.value;

      let html = '<option value="ALL">⚡ ALL HOSTS (Fleet-Wide Simultaneous Execution)</option>';
      html += '<option value="ALL_WINDOWS">🪟 All Windows Hosts</option>';
      html += '<option value="ALL_LINUX">🐧 All Linux Hosts</option>';

      agents.forEach(a => {
        html += '<option value="' + escapeHtml(a.id) + '">' + escapeHtml(a.hostname || a.id) + ' (' + escapeHtml(a.primary_ip || '127.0.0.1') + ') - ' + escapeHtml(a.platform || a.os || 'Host') + '</option>';
      });

      targetSel.innerHTML = html;
      if (curVal) targetSel.value = curVal;
    }

    // Live Persona Updating (Persists & Never Reverts)
    async function changeAgentPersona(agentId, persona) {
      try {
        const res = await fetch('/api/v1/controller/set_persona', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id: agentId, persona: persona })
        });
        if (res.ok) {
          showToast('Updated agent persona to ' + persona);
          if (fleetAgentsCache) {
            const ag = fleetAgentsCache.find(x => x.id === agentId);
            if (ag) ag.assigned_persona = persona;
          }
        } else {
          showToast('Failed to update persona');
        }
      } catch (e) {
        showToast('Error updating persona: ' + e);
      }
    }

    async function applyBulkPersona() {
      const sel = document.getElementById('bulkPersonaSelect');
      if (!sel) return;
      const p = sel.value;
      try {
        const res = await fetch('/api/v1/controller/set_persona', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id: 'ALL', persona: p })
        });
        if (res.ok) {
          showToast('Updated persona for ALL agents to ' + p);
          if (fleetAgentsCache) {
            fleetAgentsCache.forEach(a => { a.assigned_persona = p; });
            renderOverviewFleetTable(fleetAgentsCache);
            renderFleetDetailTable(fleetAgentsCache);
          }
        } else {
          showToast('Failed to apply fleet-wide persona');
        }
      } catch (e) {
        showToast('Error: ' + e);
      }
    }

    // =========================================================================
    // FLEET-WIDE COMMAND EXECUTION (WITH STEALTH HISTORY SUPPRESSION)
    // =========================================================================

    function setFleetCmd(cmd) {
      const el = document.getElementById('fleetCmdInput');
      if (el) {
        el.value = cmd;
        el.focus();
      }
    }

    function clearFleetTerminal() {
      const el = document.getElementById('fleetCmdResults');
      if (el) el.innerHTML = '<div style="color:var(--text-muted); font-size:0.72rem;">[CONSOLE OUTPUT CLEARED]</div>';
    }

    function quickShellHost(agentId) {
      navigatePage('/fleet');
      const sel = document.getElementById('fleetCmdTarget');
      if (sel) sel.value = agentId;
      const inp = document.getElementById('fleetCmdInput');
      if (inp) inp.focus();
    }

    async function executeFleetCommand() {
      const input = document.getElementById('fleetCmdInput');
      const targetSel = document.getElementById('fleetCmdTarget');
      const term = document.getElementById('fleetCmdResults');
      if (!input || !term) return;
      const cmd = input.value.trim();
      if (!cmd) return;
      const target = targetSel ? targetSel.value : 'ALL';

      const timestamp = new Date().toISOString().replace('T', ' ').split('.')[0];
      const blockId = 'cmd-block-' + Date.now();

      term.innerHTML += '<div class="cmd-run-block" id="' + blockId + '"><div class="cmd-run-header">[' + timestamp + ' UTC] &gt; <strong>' + escapeHtml(cmd) + '</strong> &nbsp;[Target: ' + escapeHtml(target) + '] [Stealth History Suppression: Active]</div><div class="cmd-run-body" id="' + blockId + '-body">Dispatching command across target host(s)...</div></div>';
      term.scrollTop = term.scrollHeight;
      input.value = '';

      try {
        let payload = { command: cmd, target_agent_id: target, agent_id: target };
        if (target === 'ALL' || target === 'ALL_WINDOWS' || target === 'ALL_LINUX') {
          payload.target_scope = target;
        }

        const res = await fetch('/api/v1/controller/command', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await res.json();
        const bodyEl = document.getElementById(blockId + '-body');

        if (res.ok) {
          if (data.task_ids && Array.isArray(data.task_ids)) {
            if (bodyEl) bodyEl.innerHTML = '<div style="color:var(--accent-emerald);">Dispatched to ' + data.task_ids.length + ' online host(s) simultaneously. Waiting for results...</div>';
            pollBatchResults(data.task_ids, blockId);
          } else if (data.task_id) {
            if (bodyEl) bodyEl.innerHTML = '<div style="color:var(--accent-emerald);">Dispatched to host ' + escapeHtml(data.target_agent_id || target) + ' (Task ID: ' + escapeHtml(data.task_id) + '). Awaiting output...</div>';
            pollSingleResult(data.task_id, blockId);
          } else {
            if (bodyEl) bodyEl.innerText = data.message || 'Queued';
          }
        } else {
          if (bodyEl) bodyEl.innerHTML = '<div style="color:var(--accent-crimson);">[DISPATCH ERROR] ' + escapeHtml(data.error || 'Failed to dispatch command') + '</div>';
        }
      } catch (e) {
        const bodyEl = document.getElementById(blockId + '-body');
        if (bodyEl) bodyEl.innerHTML = '<div style="color:var(--accent-crimson);">[NETWORK ERROR] ' + escapeHtml(String(e)) + '</div>';
      }
      term.scrollTop = term.scrollHeight;
    }

    async function pollSingleResult(taskId, blockId) {
      const bodyEl = document.getElementById(blockId + '-body');
      for (let i = 0; i < 8; i++) {
        await new Promise(r => setTimeout(r, 600));
        try {
          const res = await fetch('/api/v1/controller/command/status?task_id=' + encodeURIComponent(taskId));
          if (res.ok) {
            const data = await res.json();
            if (data.status === 'success' || data.status === 'completed' || data.exit_code !== undefined) {
              let out = data.stdout || '';
              if (data.stderr) out += (out ? String.fromCharCode(10) : '') + '[STDERR] ' + data.stderr;
              if (!out && data.status === 'success') out = '[Command completed with Exit Code 0 - No stdout output]';
              if (bodyEl) {
                bodyEl.innerHTML = '<div style="color:var(--accent-emerald); margin-bottom:4px;">[Exit Code ' + (data.exit_code ?? 0) + ' | ' + escapeHtml(data.agent_id || 'Host') + ']</div><pre style="font-family:inherit; white-space:pre-wrap;">' + escapeHtml(out) + '</pre>';
              }
              const term = document.getElementById('fleetCmdResults');
              if (term) term.scrollTop = term.scrollHeight;
              return;
            }
          }
        } catch (e) {}
      }
    }

    async function pollBatchResults(taskIds, blockId) {
      const bodyEl = document.getElementById(blockId + '-body');
      for (let i = 0; i < 8; i++) {
        await new Promise(r => setTimeout(r, 800));
        try {
          const promises = taskIds.map(tId => fetch('/api/v1/controller/command/status?task_id=' + encodeURIComponent(tId)).then(r => r.json()).catch(() => null));
          const results = await Promise.all(promises);
          const completed = results.filter(r => r && (r.status === 'success' || r.status === 'completed' || r.exit_code !== undefined));

          if (completed.length > 0 && bodyEl) {
            let html = '<div style="color:var(--accent-emerald); margin-bottom:6px;">Received responses from ' + completed.length + ' / ' + taskIds.length + ' hosts:</div>';
            completed.forEach(c => {
              let out = c.stdout || '';
              if (c.stderr) out += (out ? String.fromCharCode(10) : '') + '[STDERR] ' + c.stderr;
              if (!out) out = '[Exit Code ' + (c.exit_code ?? 0) + ' - No stdout]';
              html += '<div style="margin-top:6px; padding:4px 6px; background:rgba(255,255,255,0.03); border-radius:4px;"><strong style="color:var(--accent-emerald);">' + escapeHtml(c.agent_id || 'Host') + '</strong> (Exit: ' + (c.exit_code ?? 0) + '):<pre style="font-family:inherit; white-space:pre-wrap; margin-top:2px;">' + escapeHtml(out) + '</pre></div>';
            });
            bodyEl.innerHTML = html;
            const term = document.getElementById('fleetCmdResults');
            if (term) term.scrollTop = term.scrollHeight;
            if (completed.length === taskIds.length) return;
          }
        } catch (e) {}
      }
    }

    function copyIp(ip) {
      if (!ip) return;
      navigator.clipboard.writeText(ip).then(() => {
        showToast('IP copied: ' + ip);
      }).catch(() => {
        showToast('IP: ' + ip);
      });
    }

    // =========================================================================
    // EVENTS & AUDIT LOGS
    // =========================================================================

    // Protocol Filtering and Instant Search for Activity Streams
    let rawOverviewEvents = [];
    let overviewFilterProto = 'ALL';
    let overviewFilterQuery = '';

    function setOverviewFilter(proto, btn) {
      overviewFilterProto = proto;
      if (btn && btn.parentElement) {
        btn.parentElement.querySelectorAll('.filter-pill').forEach(p => p.classList.remove('active'));
        btn.classList.add('active');
      }
      renderFilteredOverviewEvents();
    }

    function handleOverviewSearch(val) {
      overviewFilterQuery = (val || '').toLowerCase().trim();
      renderFilteredOverviewEvents();
    }

    function renderFilteredOverviewEvents() {
      const tbody = document.getElementById('overviewEventsBody');
      if (!tbody) return;
      let events = rawOverviewEvents || [];
      if (overviewFilterProto !== 'ALL') {
        events = events.filter(ev => (ev.protocol || '').toUpperCase().includes(overviewFilterProto));
      }
      if (overviewFilterQuery) {
        events = events.filter(ev => {
          const text = ((ev.agent_id || '') + ' ' + (ev.protocol || '') + ' ' + (ev.action || ev.details || '') + ' ' + (ev.status || '')).toLowerCase();
          return text.includes(overviewFilterQuery);
        });
      }
      if (events.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; color:var(--text-muted); padding:2rem;">No emulation events match the filter criteria.</td></tr>';
        return;
      }
      let html = '';
      events.slice(0, 10).forEach(ev => {
        const isSuccess = ev.status === 'success';
        const badgeClass = isSuccess ? 'status-running' : 'status-stopped';
        const timeStr = ev.timestamp ? ev.timestamp.replace('T', ' ').split('.')[0] : '-';
        html += '<tr>';
        html += '<td class="mono">' + timeStr + '</td>';
        html += '<td><strong>' + escapeHtml(ev.agent_id || 'MANAGER') + '</strong></td>';
        html += '<td><span class="status-badge status-standby mono">' + escapeHtml(ev.protocol || 'SYS') + '</span></td>';
        html += '<td>' + escapeHtml(ev.action || ev.details || '-') + '</td>';
        html += '<td><span class="status-badge ' + badgeClass + '">' + escapeHtml(ev.status || 'info') + '</span></td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    }

    async function loadRecentEvents() {
      try {
        const res = await fetch('/api/v1/controller/events?limit=30');
        if (!res.ok) return;
        const data = await res.json();
        const events = data.events || [];
        rawOverviewEvents = events;
        if (events.length > cumulativeActions) {
          cumulativeActions = events.length;
        }
        const countBadge = document.getElementById('navEventCount');
        if (countBadge) countBadge.innerText = events.length;
        renderFilteredOverviewEvents();
      } catch (e) {
        console.error('Failed to load recent events:', e);
      }
    }

    let rawAuditEvents = [];
    let auditFilterProto = 'ALL';
    let auditFilterQuery = '';

    function setAuditFilter(proto, btn) {
      auditFilterProto = proto;
      if (btn && btn.parentElement) {
        btn.parentElement.querySelectorAll('.filter-pill').forEach(p => p.classList.remove('active'));
        btn.classList.add('active');
      }
      renderFilteredAuditEvents();
    }

    function handleAuditSearch(val) {
      auditFilterQuery = (val || '').toLowerCase().trim();
      renderFilteredAuditEvents();
    }

    function renderFilteredAuditEvents() {
      const tbody = document.getElementById('fullAuditBody');
      if (!tbody) return;
      let events = rawAuditEvents || [];
      if (auditFilterProto !== 'ALL') {
        events = events.filter(ev => (ev.protocol || '').toUpperCase().includes(auditFilterProto));
      }
      if (auditFilterQuery) {
        events = events.filter(ev => {
          const text = ((ev.agent_id || '') + ' ' + (ev.protocol || '') + ' ' + (ev.action || ev.details || '') + ' ' + (ev.status || '')).toLowerCase();
          return text.includes(auditFilterQuery);
        });
      }
      if (events.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; color:var(--text-muted); padding:2rem;">No audit logs match criteria.</td></tr>';
        return;
      }
      let html = '';
      events.forEach(ev => {
        const isSuccess = ev.status === 'success';
        const badgeClass = isSuccess ? 'status-running' : 'status-stopped';
        html += '<tr>';
        html += '<td class="mono">' + escapeHtml(ev.timestamp || '-') + '</td>';
        html += '<td><strong>' + escapeHtml(ev.agent_id || 'MANAGER') + '</strong></td>';
        html += '<td><span class="status-badge status-standby mono">' + escapeHtml(ev.protocol || 'SYS') + '</span></td>';
        html += '<td>' + escapeHtml(ev.action || ev.details || '-') + '</td>';
        html += '<td><span class="status-badge ' + badgeClass + '">' + escapeHtml(ev.status || 'info') + '</span></td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    }

    async function loadFullAuditLog() {
      try {
        const res = await fetch('/api/v1/controller/events?limit=250');
        if (!res.ok) return;
        const data = await res.json();
        rawAuditEvents = data.events || [];
        if (rawAuditEvents.length > cumulativeActions) {
          cumulativeActions = rawAuditEvents.length;
        }
        renderFilteredAuditEvents();
      } catch (e) {
        console.error('Failed to load audit logs:', e);
      }
    }

    async function exportAuditJson() {
      try {
        const res = await fetch('/api/v1/controller/events?limit=1000');
        const data = await res.json();
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'range_audit_log.json';
        a.click();
        URL.revokeObjectURL(url);
        showToast('Audit log exported successfully');
      } catch (e) {
        showToast('Export failed: ' + e);
      }
    }

    // =========================================================================
    // WEB CORPUS & WORDLIST (SIMPLIFIED HTTP/HTTPS ONE SITE PER LINE)
    // =========================================================================

    async function loadCorpusData() {
      try {
        const [cRes, wRes] = await Promise.all([
          fetch('/api/v1/controller/corpus'),
          fetch('/api/v1/controller/wordlist')
        ]);
        if (cRes.ok) {
          const cData = await cRes.json();
          const cEl = document.getElementById('corpusEditor');
          if (cEl) {
            if (cData.urls) {
              cEl.value = cData.urls;
            } else {
              const all = (cData.intranet_portals || []).concat(cData.internet_sites || []);
              cEl.value = all.join(String.fromCharCode(10));
            }
          }
        }
        if (wRes.ok) {
          const wData = await wRes.json().catch(() => null);
          const wEl = document.getElementById('wordlistEditor');
          if (wEl) {
            if (wData && typeof wData.words === 'string') {
              wEl.value = wData.words;
            } else if (typeof wData === 'string') {
              wEl.value = wData;
            }
          }
        }
      } catch (e) {
        showToast('Failed to load corpus data: ' + e);
      }
    }

    async function saveWebCorpus() {
      const el = document.getElementById('corpusEditor');
      if (!el) return;
      const rawText = el.value.trim();
      const lines = rawText.split(String.fromCharCode(10)).map(l => l.trim()).filter(l => l.length > 0);

      // Validate http:// or https://
      for (const line of lines) {
        if (!line.startsWith('http://') && !line.startsWith('https://')) {
          alert('Invalid site entry: "' + line + '". Each line must start with http:// or https://. Strictly one site per line.');
          return;
        }
      }

      try {
        const res = await fetch('/api/v1/controller/corpus', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ urls: lines.join(String.fromCharCode(10)) })
        });
        if (res.ok) {
          showToast('Web corpus saved (' + lines.length + ' sites) & broadcast live');
        } else {
          showToast('Failed to save corpus');
        }
      } catch (e) {
        showToast('Error saving corpus: ' + e);
      }
    }

    async function saveWordlist() {
      const el = document.getElementById('wordlistEditor');
      if (!el) return;
      try {
        const res = await fetch('/api/v1/controller/wordlist', {
          method: 'POST',
          headers: { 'Content-Type': 'text/plain' },
          body: el.value
        });
        if (res.ok) {
          showToast('Wordlist saved successfully');
        } else {
          showToast('Failed to save wordlist');
        }
      } catch (e) {
        showToast('Error saving wordlist: ' + e);
      }
    }

    // =========================================================================
    // SIMPLIFIED NETWORK TOPOLOGY
    // =========================================================================

    let topologyNodes = [];

    async function loadTopologyData() {
      try {
        const [tRes, sRes] = await Promise.all([
          fetch('/api/v1/controller/topology'),
          fetch('/api/v1/controller/schedules'),
          loadFleetAgents()
        ]);
        if (tRes.ok) {
          const tData = await tRes.json();
          const tEl = document.getElementById('topologyEditor');
          if (tEl) tEl.value = JSON.stringify(tData, null, 2);

          if (Array.isArray(tData)) {
            topologyNodes = tData;
          } else if (tData && Array.isArray(tData.routers)) {
            topologyNodes = tData.routers.map(r => ({
              name: r.hostname || 'Gateway',
              ip: r.control_ip || '10.0.0.1',
              subnet: r.external_ip || '10.0.0.0/16',
              description: r.platform_role || 'Routing Gateway'
            }));
          } else {
            topologyNodes = [
              { name: 'Edge Gateway', ip: '10.0.0.1', subnet: '10.0.0.0/16', description: 'Uplink & Core DNS' },
              { name: 'Workstations Gateway', ip: '10.0.10.1', subnet: '10.0.10.0/24', description: 'User Workstations' }
            ];
          }
          renderTopologyTable();
          renderTopologyMap();
        }
        if (sRes.ok) {
          const sData = await sRes.json();
          const sEl = document.getElementById('schedulesEditor');
          if (sEl) sEl.value = JSON.stringify(sData, null, 2);
        }
      } catch (e) {
        console.error('Failed to load topology:', e);
      }
    }

    // Live SVG Network Topology Graph
    function renderTopologyMap() {
      const svg = document.getElementById('topologySvgMap');
      if (!svg) return;

      const nodes = (topologyNodes && topologyNodes.length > 0) ? topologyNodes : [
        { name: 'Workstations Gateway', ip: '10.0.10.1', subnet: '10.0.10.0/24' }
      ];
      const agents = fleetAgentsCache || [];

      let svgHtml = '<defs>';
      svgHtml += '<linearGradient id="coreGrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0%" stop-color="#0284c7"/><stop offset="100%" stop-color="#0369a1"/></linearGradient>';
      svgHtml += '<linearGradient id="nodeGrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0%" stop-color="#1b2838"/><stop offset="100%" stop-color="#0d1929"/></linearGradient>';
      svgHtml += '<linearGradient id="agentGrad" x1="0" y1="0" x2="1" y2="1"><stop offset="0%" stop-color="#141c2b"/><stop offset="100%" stop-color="#0b121e"/></linearGradient>';
      svgHtml += '<filter id="coreGlow" x="-20%" y="-20%" width="140%" height="140%"><feGaussianBlur stdDeviation="3" result="blur"/><feComposite in="SourceGraphic" in2="blur" operator="over"/></filter>';
      svgHtml += '</defs>';

      // Draw Range Edge Core Gateway at center top: (450, 32)
      const coreX = 450, coreY = 32;
      svgHtml += '<rect x="340" y="14" width="220" height="38" rx="6" fill="url(#coreGrad)" stroke="#38bdf8" stroke-width="1.5" filter="url(#coreGlow)"/>';
      svgHtml += '<text x="450" y="30" fill="#ffffff" font-family="Inter, sans-serif" font-size="11" font-weight="700" text-anchor="middle">EDGE UPLINK &amp; CORE ROUTER</text>';
      svgHtml += '<text x="450" y="43" fill="#bae6fd" font-family="JetBrains Mono, monospace" font-size="9" text-anchor="middle">10.0.0.1 // RANGE-CORE-GW</text>';

      // Middle Row: Gateway Subnets
      const gwCount = nodes.length;
      const gwSpacing = Math.min(260, 820 / (gwCount + 1));
      const gwStartX = 450 - ((gwCount - 1) * gwSpacing / 2);
      const gwY = 108;

      const gwCoords = [];
      nodes.forEach((n, idx) => {
        const x = Math.round(gwStartX + idx * gwSpacing);
        gwCoords.push({ x: x, y: gwY, subnet: n.subnet || '' });

        // Curved line from Core to Gateway
        svgHtml += '<path d="M ' + coreX + ' 52 C ' + coreX + ' 78, ' + x + ' 78, ' + x + ' 92" fill="none" stroke="#0284c7" stroke-width="1.5" stroke-dasharray="3,3"/>';

        // Gateway Node Box
        svgHtml += '<rect x="' + (x - 85) + '" y="92" width="170" height="38" rx="6" fill="url(#nodeGrad)" stroke="#1d3c6a" stroke-width="1.5"/>';
        svgHtml += '<text x="' + x + '" y="108" fill="#f0fdfa" font-family="Inter, sans-serif" font-size="10" font-weight="600" text-anchor="middle">' + escapeHtml(n.name || 'Gateway') + '</text>';
        svgHtml += '<text x="' + x + '" y="121" fill="#38bdf8" font-family="JetBrains Mono, monospace" font-size="9" text-anchor="middle">' + escapeHtml(n.subnet || n.ip || '') + '</text>';
      });

      // Bottom Row: Connected Fleet Host Agents
      if (agents.length === 0) {
        svgHtml += '<text x="450" y="195" fill="#7dd3fc" font-family="Inter, sans-serif" font-size="11" text-anchor="middle" opacity="0.75">Awaiting agent connection to populate live endpoint topology...</text>';
      } else {
        const agentCount = agents.length;
        const agSpacing = Math.min(180, 840 / agentCount);
        const agStartX = 450 - ((agentCount - 1) * agSpacing / 2);

        agents.forEach((a, idx) => {
          const ax = Math.round(agStartX + idx * agSpacing);
          const targetGw = gwCoords[idx % gwCoords.length] || { x: coreX, y: gwY };

          const isOnline = (a.status || '').toLowerCase() === 'online';
          const strokeColor = isOnline ? '#10b981' : '#64748b';
          svgHtml += '<path d="M ' + targetGw.x + ' 130 C ' + targetGw.x + ' 158, ' + ax + ' 158, ' + ax + ' 178" fill="none" stroke="' + strokeColor + '" stroke-width="1.2"/>';

          // Host Node Card
          svgHtml += '<g style="cursor:pointer;" onclick="inspectAgent(\'' + escapeHtml(a.id) + '\')" title="Click to Inspect ' + escapeHtml(a.hostname || a.id) + '">';
          svgHtml += '<rect x="' + (ax - 68) + '" y="178" width="136" height="42" rx="5" fill="url(#agentGrad)" stroke="' + (isOnline ? '#059669' : '#334155') + '" stroke-width="1.2"/>';
          svgHtml += '<circle cx="' + (ax - 52) + '" cy="193" r="4" fill="' + (isOnline ? '#10b981' : '#94a3b8') + '"/>';
          svgHtml += '<text x="' + (ax + 2) + '" y="196" fill="#f8fafc" font-family="Inter, sans-serif" font-size="9.5" font-weight="600" text-anchor="middle">' + escapeHtml((a.hostname || a.id).slice(0, 14)) + '</text>';
          svgHtml += '<text x="' + ax + '" y="210" fill="#94a3b8" font-family="JetBrains Mono, monospace" font-size="8.5" text-anchor="middle">' + escapeHtml(a.primary_ip || '127.0.0.1') + '</text>';
          svgHtml += '</g>';
        });
      }

      svg.innerHTML = svgHtml;
    }

    function renderTopologyTable() {
      const tbody = document.getElementById('topologyTableBody');
      if (!tbody) return;
      if (!topologyNodes || topologyNodes.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; color:var(--text-muted); padding:1.5rem;">No gateway nodes configured. Click "+ Add Gateway Node" to add routing nodes.</td></tr>';
        return;
      }
      let html = '';
      topologyNodes.forEach((node, idx) => {
        html += '<tr>';
        html += '<td><input type="text" class="form-input" style="padding:0.35rem 0.5rem; font-size:0.75rem;" value="' + escapeHtml(node.name || '') + '" placeholder="e.g. Edge Gateway" data-idx="' + idx + '" data-field="name" onchange="handleTopologyField(this)"></td>';
        html += '<td><input type="text" class="form-input form-input-mono" style="padding:0.35rem 0.5rem; font-size:0.75rem;" value="' + escapeHtml(node.ip || '') + '" placeholder="10.0.0.1" data-idx="' + idx + '" data-field="ip" onchange="handleTopologyField(this)"></td>';
        html += '<td><input type="text" class="form-input form-input-mono" style="padding:0.35rem 0.5rem; font-size:0.75rem;" value="' + escapeHtml(node.subnet || '') + '" placeholder="10.0.0.0/16" data-idx="' + idx + '" data-field="subnet" onchange="handleTopologyField(this)"></td>';
        html += '<td><input type="text" class="form-input" style="padding:0.35rem 0.5rem; font-size:0.75rem;" value="' + escapeHtml(node.description || '') + '" placeholder="Role / Description" data-idx="' + idx + '" data-field="description" onchange="handleTopologyField(this)"></td>';
        html += '<td style="text-align:right;"><button class="btn btn-danger" style="padding:0.25rem 0.55rem; font-size:0.70rem;" onclick="removeTopologyNode(' + idx + ')">Remove</button></td>';
        html += '</tr>';
      });
      tbody.innerHTML = html;
    }

    function handleTopologyField(el) {
      const idx = parseInt(el.dataset.idx, 10);
      const field = el.dataset.field;
      if (!isNaN(idx) && topologyNodes[idx] && field) {
        topologyNodes[idx][field] = el.value.trim();
        const tEl = document.getElementById('topologyEditor');
        if (tEl) tEl.value = JSON.stringify(topologyNodes, null, 2);
        renderTopologyMap();
      }
    }

    function addTopologyNode() {
      topologyNodes.push({
        name: 'Gateway Node ' + (topologyNodes.length + 1),
        ip: '10.0.' + ((topologyNodes.length + 1) * 10) + '.1',
        subnet: '10.0.' + ((topologyNodes.length + 1) * 10) + '.0/24',
        description: 'Single-network routing node'
      });
      renderTopologyTable();
      renderTopologyMap();
      const tEl = document.getElementById('topologyEditor');
      if (tEl) tEl.value = JSON.stringify(topologyNodes, null, 2);
    }

    function removeTopologyNode(idx) {
      topologyNodes.splice(idx, 1);
      renderTopologyTable();
      renderTopologyMap();
      const tEl = document.getElementById('topologyEditor');
      if (tEl) tEl.value = JSON.stringify(topologyNodes, null, 2);
    }

    function toggleRawTopology() {
      const wrap = document.getElementById('rawTopologyWrap');
      if (wrap) wrap.style.display = (wrap.style.display === 'none') ? 'block' : 'none';
    }

    async function saveTopologyFromTable() {
      try {
        const res = await fetch('/api/v1/controller/topology', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(topologyNodes)
        });
        if (res.ok) {
          showToast('Topology saved (' + topologyNodes.length + ' gateway nodes)');
          loadTopologyData();
        } else {
          showToast('Failed to save topology');
        }
      } catch (e) {
        showToast('Error saving topology: ' + e);
      }
    }

    async function saveRawTopology() {
      const el = document.getElementById('topologyEditor');
      if (!el) return;
      try {
        const parsed = JSON.parse(el.value);
        const res = await fetch('/api/v1/controller/topology', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(parsed)
        });
        if (res.ok) {
          showToast('Raw topology saved');
          loadTopologyData();
        } else {
          showToast('Failed to save topology');
        }
      } catch (e) {
        alert('Invalid JSON: ' + e.message);
      }
    }

    async function saveSchedules() {
      const el = document.getElementById('schedulesEditor');
      if (!el) return;
      const val = el.value.trim();
      if (!val || val === '[]' || val === '{}') {
        showToast('Schedule configuration is optional. No schedules set.');
        return;
      }
      try {
        const parsed = JSON.parse(val);
        const res = await fetch('/api/v1/controller/schedules', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(parsed)
        });
        if (res.ok) {
          showToast('Emulation schedules saved (Optional)');
        } else {
          showToast('Failed to save schedules', true);
        }
      } catch (e) {
        showToast('Optional schedule note: ' + e.message, true);
      }
    }

    // =========================================================================
    // ADMIN AUTHENTICATION
    // =========================================================================

    async function submitAdminPassword() {
      const oldPass = document.getElementById('oldAdminPass').value;
      const newUser = document.getElementById('newAdminUser').value.trim();
      const newPass = document.getElementById('newAdminPass').value;

      if (!oldPass) {
        alert('Current password is required');
        return;
      }
      if (newPass && newPass.length < 8) {
        alert('New password must be at least 8 characters');
        return;
      }
      try {
        const res = await fetch('/api/v1/auth/change_credentials', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            current_password: oldPass,
            new_username: newUser,
            new_password: newPass
          })
        });
        const data = await res.json();
        if (res.ok) {
          showToast('Admin credentials updated successfully');
          if (newUser) {
            document.getElementById('sidebarUserLabel').innerText = newUser;
          }
          closeModal('adminModal');
        } else {
          alert('Update failed: ' + (data.error || 'Unknown error'));
        }
      } catch (e) {
        alert('Network error: ' + e);
      }
    }

    async function handleLogout() {
      try {
        await fetch('/api/v1/auth/logout', { method: 'POST' });
        showToast('Signed out of session');
        setTimeout(() => { location.reload(); }, 500);
      } catch (e) {
        location.reload();
      }
    }

    function escapeHtml(unsafe) {
      if (unsafe === null || unsafe === undefined) return '';
      return String(unsafe)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
    }

    // Polling Loop
    async function poll() {
      if (isRequestBusy) return;
      isRequestBusy = true;
      try {
        await loadRangeDetails();
        if (currentPath === '/' || currentPath.includes('overview')) {
          await Promise.all([loadFleetAgents(), loadRecentEvents()]);
        } else if (currentPath.includes('fleet') || currentPath.includes('configure')) {
          await loadFleetAgents();
        } else if (currentPath.includes('topology')) {
          await loadTopologyData();
        } else if (currentPath.includes('activity')) {
          await loadFullAuditLog();
        }
      } catch (e) {
      } finally {
        isRequestBusy = false;
      }
    }

    // Bootstrap
    initTheme();
    initSidebar();
    updateClocks();
    setInterval(updateClocks, 1000);
    setInterval(updateSessionTicker, 1000);
    applyActivePage(window.location.pathname);
    setInterval(poll, 3000);
  </script>
</body>
</html>
`

// ServeDashboard handles HTTP GET for the web operations dashboard.
func ServeDashboard(w http.ResponseWriter, r *http.Request, s *Server) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(DashboardHTML))
}
