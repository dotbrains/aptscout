// aptscout SPA
(function () {
  'use strict';

  const $ = (sel) => document.querySelector(sel);
  const app = $('#app');
  let filters = { property: null, beds: null, baths: null, min_price: null, max_price: null, plan: null, renovated: null, available_from: null, available_to: null, sort: 'price', order: 'asc' };

  // Router
  function route() {
    const hash = location.hash || '#/';
    if (hash === '#/' || hash === '#') renderPropertyPicker();
    else if (hash.startsWith('#/property/')) {
      const prop = hash.substring(11);
      filters.property = prop;
      renderDashboard();
    }
    else if (hash === '#/floor-plans') renderFloorPlans();
    else if (hash.startsWith('#/unit/')) {
      const parts = hash.substring(7).split('/');
      renderUnit(parts[0], parts[1]);
    }
    else renderPropertyPicker();
    updateNavLinks();
  }

  function updateNavLinks() {
    document.querySelectorAll('.nav-link').forEach(l => {
      l.classList.toggle('active', l.getAttribute('href') === (location.hash || '#/'));
    });
  }

  // API
  async function api(path, opts) {
    const res = await fetch('/api/' + path, opts);
    return res.json();
  }

  // Toast
  function toast(msg) {
    const t = $('#toast');
    t.textContent = msg;
    t.classList.add('show');
    setTimeout(() => t.classList.remove('show'), 3000);
  }

  // Scrape button
  function showLoading(text, sub) {
    const overlay = document.createElement('div');
    overlay.className = 'loading-overlay';
    overlay.id = 'loading-overlay';
    overlay.innerHTML = `
      <div class="loading-spinner"></div>
      <div class="loading-text">${text || 'Scraping...'}</div>
      <div class="loading-sub" id="loading-sub">${sub || 'Fetching apartments from all properties'}</div>
    `;
    document.body.appendChild(overlay);
  }

  function hideLoading() {
    document.getElementById('loading-overlay')?.remove();
  }

  function updateLoadingSub(text) {
    const el = document.getElementById('loading-sub');
    if (el) el.textContent = text;
  }

  async function doScrape() {
    const btn = $('#btn-scrape');
    btn.classList.add('loading');
    btn.innerHTML = '<i data-lucide="loader-2" class="icon spin"></i> Scraping...';
    showLoading('Scraping apartments...', 'Fetching data from all properties');
    try {
      const r = await api('scrape', { method: 'POST' });
      const parts = [`${r.UnitsFound} units`];
      if (r.UnitsNew) parts.push(`${r.UnitsNew} new`);
      if (r.UnitsChanged) parts.push(`${r.UnitsChanged} changed`);
      hideLoading();
      toast(`Scrape complete: ${parts.join(', ')}`);
      route();
    } catch (e) {
      hideLoading();
      toast('Scrape failed: ' + e.message);
    } finally {
      btn.classList.remove('loading');
      btn.innerHTML = '<i data-lucide="refresh-cw" class="icon"></i> Scrape';
      lucide.createIcons();
      loadLastScraped();
    }
  }

  async function loadLastScraped() {
    try {
      const stats = await api('stats');
      const el = $('#last-scraped');
      if (stats.last_scrape) {
        const d = new Date(stats.last_scrape);
        el.textContent = 'Last scraped: ' + timeAgo(d);
      }
    } catch (_) {}
  }

  // Property picker (home page)
  async function renderPropertyPicker() {
    const stats = await api('stats');
    // Get per-property stats
    const dcStats = await api('stats?property=desert-club');
    const haStats = await api('stats?property=hideaway');

    const properties = [
      { id: 'desert-club', name: 'Desert Club Apartments', icon: '🌵', location: 'Phoenix, AZ', stats: dcStats },
      { id: 'hideaway', name: 'Hideaway North Scottsdale', icon: '🏜️', location: 'Scottsdale, AZ', stats: haStats },
    ];

    app.innerHTML = `
      <div class="property-picker">
        <div class="property-picker-title"><i data-lucide="building-2" class="icon-lg"></i> Select a Property</div>
        <div class="property-picker-subtitle">Choose an apartment complex to browse available units</div>
        <div class="property-grid">
          ${properties.map(p => `
            <div class="property-card" data-nav-property="${p.id}">
              <div class="property-card-name">${p.icon} ${p.name}</div>
              <div class="property-card-id">${p.id} · ${p.location}</div>
              <div class="property-card-stats">
                <i data-lucide="door-open" class="icon-sm"></i> ${p.stats.available || 0} units available ·
                <i data-lucide="layout-grid" class="icon-sm"></i> ${p.stats.floor_plans || 0} floor plans
              </div>
            </div>
          `).join('')}
        </div>
        <div style="margin-top:24px;color:var(--muted);font-size:13px">
          <i data-lucide="info" class="icon-sm"></i>
          Total across all properties: ${stats.available || 0} units, ${stats.floor_plans || 0} floor plans
        </div>
      </div>
    `;

    document.querySelectorAll('[data-nav-property]').forEach(c => c.onclick = () => {
      location.hash = '#/property/' + c.dataset.navProperty;
    });

    lucide.createIcons();
  }

  // Dashboard (apartments for a selected property)
  async function renderDashboard() {
    const params = new URLSearchParams();
    if (filters.property) params.set('property', filters.property);
    if (filters.beds) params.set('beds', filters.beds);
    if (filters.baths) params.set('baths', filters.baths);
    if (filters.min_price) params.set('min_price', filters.min_price);
    if (filters.max_price) params.set('max_price', filters.max_price);
    if (filters.plan) params.set('plan', filters.plan);
    if (filters.renovated) params.set('renovated', 'true');
    if (filters.sort) params.set('sort', filters.sort);
    if (filters.order) params.set('order', filters.order);

    let apts = await api('apartments?' + params.toString());

    // Client-side date range filter
    if (filters.available_from || filters.available_to) {
      apts = apts.filter(a => {
        if (a.available_now) return true; // always include "available now"
        if (!a.available_date) return false;
        if (filters.available_from && a.available_date < filters.available_from) return false;
        if (filters.available_to && a.available_date > filters.available_to) return false;
        return true;
      });
    }
    const fpParams = new URLSearchParams();
    if (filters.property) fpParams.set('property', filters.property);
    const plans = await api('floor-plans?' + fpParams.toString());

    app.innerHTML = `
      <div class="dashboard">
        <div class="sidebar">
          <div class="filter-panel">
            <div class="filter-title"><i data-lucide="sliders-horizontal" class="icon"></i> Filters</div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="bed-double" class="icon-sm"></i> Bedrooms</span>
              <div class="filter-toggles" id="bed-toggles">
                ${[1, 2, 3].map(n => `<button class="filter-toggle${filters.beds == n ? ' active' : ''}" data-beds="${n}">${n}</button>`).join('')}
              </div>
            </div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="bath" class="icon-sm"></i> Bathrooms</span>
              <div class="filter-toggles" id="bath-toggles">
                ${[1, 2].map(n => `<button class="filter-toggle${filters.baths == n ? ' active' : ''}" data-baths="${n}">${n}</button>`).join('')}
              </div>
            </div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="dollar-sign" class="icon-sm"></i> Price Range</span>
              <div class="filter-range">
                <input type="number" class="filter-input" id="min-price" placeholder="Min" value="${filters.min_price || ''}">
                <span style="color:var(--muted)">–</span>
                <input type="number" class="filter-input" id="max-price" placeholder="Max" value="${filters.max_price || ''}">
              </div>
            </div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="layout-grid" class="icon-sm"></i> Floor Plan</span>
              <div class="filter-toggles" id="plan-toggles">
                ${plans.map(p => `<button class="filter-toggle${filters.plan === p.code ? ' active' : ''}" data-plan="${p.code}">${p.code}</button>`).join('')}
              </div>
            </div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="calendar-range" class="icon-sm"></i> Availability</span>
              <div class="filter-range">
                <input type="date" class="filter-input" id="avail-from" value="${filters.available_from || ''}">
                <span style="color:var(--muted);font-size:11px;flex-shrink:0">–</span>
                <input type="date" class="filter-input" id="avail-to" value="${filters.available_to || ''}">
              </div>
              ${filters.available_from || filters.available_to ? '<div style="margin-top:4px;font-size:11px;color:var(--accent2)"><i data-lucide="info" class="icon-sm"></i> Showing units available in this date range</div>' : ''}
            </div>
            <div class="filter-group">
              <span class="filter-label"><i data-lucide="sparkles" class="icon-sm"></i> Type</span>
              <div class="filter-toggles">
                <button class="filter-toggle${filters.renovated === null ? ' active' : ''}" data-type="all">All</button>
                <button class="filter-toggle${filters.renovated === true ? ' active' : ''}" data-type="renovated">Renovated</button>
                <button class="filter-toggle${filters.renovated === false ? ' active' : ''}" data-type="premium">Premium</button>
              </div>
            </div>
            <button class="btn btn-clear" id="btn-clear" style="width:100%;margin-top:8px"><i data-lucide="x" class="icon-sm"></i> Clear Filters</button>
          </div>
        </div>
        <div class="main-content">
          <div class="sort-bar">
            <a href="#/" class="back-link" style="margin:0;margin-right:auto"><i data-lucide="arrow-left" class="icon-sm"></i> All Properties</a>
            <span class="result-count">${apts.length} apartment${apts.length !== 1 ? 's' : ''}</span>
            <select id="sort-select">
              <option value="price-asc"${filters.sort === 'price' && filters.order === 'asc' ? ' selected' : ''}>Price: Low → High</option>
              <option value="price-desc"${filters.sort === 'price' && filters.order === 'desc' ? ' selected' : ''}>Price: High → Low</option>
              <option value="date-asc"${filters.sort === 'date' ? ' selected' : ''}>Date: Soonest</option>
              <option value="sqft-desc"${filters.sort === 'sqft' ? ' selected' : ''}>Size: Largest</option>
            </select>
          </div>
          ${apts.length === 0 ? `
            <div class="empty">
              <div class="empty-icon"><i data-lucide="${filters.available_from || filters.available_to ? 'calendar-x' : 'building-2'}" class="icon-xl"></i></div>
              <div class="empty-text">${filters.available_from || filters.available_to
                ? 'No apartments available in this date range'
                : 'No apartments found'}</div>
              <div class="empty-hint">${filters.available_from || filters.available_to
                ? `No units were found available between ${filters.available_from || 'any date'} and ${filters.available_to || 'any date'}. Try widening the range or clearing the date filter.`
                : 'Try adjusting your filters or run a scrape first.'}</div>
            </div>
          ` : `
            <div class="card-grid">
              ${apts.map(a => cardHTML(a)).join('')}
            </div>
          `}
        </div>
      </div>
    `;

    // Bind events
    document.querySelectorAll('[data-beds]').forEach(b => b.onclick = () => {
      filters.beds = filters.beds == b.dataset.beds ? null : parseInt(b.dataset.beds);
      renderDashboard();
    });
    document.querySelectorAll('[data-baths]').forEach(b => b.onclick = () => {
      filters.baths = filters.baths == b.dataset.baths ? null : parseInt(b.dataset.baths);
      renderDashboard();
    });
    document.querySelectorAll('[data-plan]').forEach(b => b.onclick = () => {
      filters.plan = filters.plan === b.dataset.plan ? null : b.dataset.plan;
      renderDashboard();
    });
    document.querySelectorAll('[data-type]').forEach(b => b.onclick = () => {
      if (b.dataset.type === 'all') filters.renovated = null;
      else if (b.dataset.type === 'renovated') filters.renovated = filters.renovated === true ? null : true;
      else filters.renovated = filters.renovated === false ? null : false;
      renderDashboard();
    });

    // Date range filter
    const fromEl = $('#avail-from'), toEl = $('#avail-to');
    if (fromEl) fromEl.onchange = () => {
      filters.available_from = fromEl.value || null;
      renderDashboard();
    };
    if (toEl) toEl.onchange = () => {
      filters.available_to = toEl.value || null;
      renderDashboard();
    };

    const minEl = $('#min-price'), maxEl = $('#max-price');
    let priceTimeout;
    const onPriceChange = () => {
      clearTimeout(priceTimeout);
      priceTimeout = setTimeout(() => {
        filters.min_price = minEl.value ? parseInt(minEl.value) : null;
        filters.max_price = maxEl.value ? parseInt(maxEl.value) : null;
        renderDashboard();
      }, 500);
    };
    if (minEl) minEl.oninput = onPriceChange;
    if (maxEl) maxEl.oninput = onPriceChange;

    const sortEl = $('#sort-select');
    if (sortEl) sortEl.onchange = () => {
      const [s, o] = sortEl.value.split('-');
      filters.sort = s;
      filters.order = o || 'asc';
      renderDashboard();
    };

    const clearBtn = $('#btn-clear');
    if (clearBtn) clearBtn.onclick = () => {
      filters = { property: filters.property, beds: null, baths: null, min_price: null, max_price: null, plan: null, renovated: null, available_from: null, available_to: null, sort: 'price', order: 'asc' };
      renderDashboard();
    };

    document.querySelectorAll('.card').forEach(c => c.onclick = () => {
      location.hash = '#/unit/' + c.dataset.property + '/' + c.dataset.unit;
    });

    lucide.createIcons();
  }

  function cardHTML(a) {
    const avail = a.available_now
      ? '<span class="card-avail now"><i data-lucide="check-circle" class="icon-sm"></i> Available Now</span>'
      : `<span class="card-avail future"><i data-lucide="calendar" class="icon-sm"></i> ${a.available_date || 'TBD'}</span>`;
    const amenities = a.amenities && a.amenities.length > 0 ? a.amenities.join(' · ') : '';
    const floor = a.floor ? ordinal(a.floor) + ' Floor' : '';
    const details = [floor, amenities].filter(Boolean).join(' · ');
    return `
      <div class="card" data-unit="${a.unit_number}" data-property="${a.property}">
        <div class="card-header">
          <span class="card-unit"><i data-lucide="door-open" class="icon-sm"></i> #${a.unit_number}</span>
          <span class="card-plan">${a.floor_plan}</span>
        </div>
        <div class="card-property"><i data-lucide="building" class="icon-sm"></i> ${a.property}</div>
        <div class="card-specs"><i data-lucide="bed-double" class="icon-sm"></i> ${a.bedrooms} bed · <i data-lucide="bath" class="icon-sm"></i> ${a.bathrooms} bath · <i data-lucide="ruler" class="icon-sm"></i> ${fmt(a.sqft)} sqft</div>
        <div class="card-price-row">
          <span class="card-price"><i data-lucide="dollar-sign" class="icon-sm"></i>${fmt(a.price)}/mo</span>
          ${avail}
        </div>
        ${details ? `<div class="card-details"><i data-lucide="map-pin" class="icon-sm"></i> <span>${details}</span></div>` : ''}
        ${a.deposit ? `<div class="card-details"><i data-lucide="shield" class="icon-sm"></i> <span>Deposit: $${fmt(a.deposit)}</span></div>` : ''}
      </div>
    `;
  }

  // Unit detail
  async function renderUnit(property, unit) {
    const data = await api('apartments/' + property + '/' + unit);
    const a = data.apartment;
    const history = data.price_history || [];

    const avail = a.available_now ? 'Available Now' : (a.available_date || 'TBD');
    const amenities = a.amenities && a.amenities.length > 0 ? a.amenities.join(', ') : 'None';

    const propertyInfo = {
      'desert-club': { name: 'Desert Club', floorPlanURL: plan => `https://arizona.weidner.com/apartments/az/phoenix/desert-club0/floorplans/${plan.toLowerCase()}` },
      'hideaway':    { name: 'Hideaway',    floorPlanURL: () => 'https://www.hideawaynorthscottsdale.com/floorplans' },
    };
    const info = propertyInfo[a.property] || { name: a.property, floorPlanURL: () => '#' };
    const floorPlanURL = info.floorPlanURL(a.floor_plan);

    app.innerHTML = `
      <div class="unit-detail">
        <a href="#/" class="back-link"><i data-lucide="arrow-left" class="icon-sm"></i> Back to apartments</a>
        <div class="unit-header">
          <div class="unit-title">#${a.unit_number} — ${a.floor_plan}</div>
          <div class="unit-subtitle">${a.bedrooms} bed · ${a.bathrooms} bath · ${fmt(a.sqft)} sq ft · ${a.is_renovated ? 'Renovated' : 'Premium'}</div>
        </div>
        <div class="unit-meta">
          <div class="meta-card"><div class="meta-label"><i data-lucide="dollar-sign" class="icon-sm"></i> Price</div><div class="meta-value price">$${fmt(a.price)}/mo</div></div>
          <div class="meta-card"><div class="meta-label"><i data-lucide="calendar" class="icon-sm"></i> Available</div><div class="meta-value">${avail}</div></div>
          <div class="meta-card"><div class="meta-label"><i data-lucide="layers" class="icon-sm"></i> Floor</div><div class="meta-value">${a.floor ? ordinal(a.floor) : '—'}</div></div>
          <div class="meta-card"><div class="meta-label"><i data-lucide="shield" class="icon-sm"></i> Deposit</div><div class="meta-value">$${fmt(a.deposit)}</div></div>
          <div class="meta-card"><div class="meta-label"><i data-lucide="list" class="icon-sm"></i> Amenities</div><div class="meta-value" style="font-size:14px">${amenities}</div></div>
          <div class="meta-card"><div class="meta-label"><i data-lucide="external-link" class="icon-sm"></i> Floor Plan</div><div class="meta-value"><a href="${floorPlanURL}" target="_blank">${a.floor_plan} on ${info.name} <i data-lucide="arrow-up-right" class="icon-sm"></i></a></div></div>
        </div>
        ${history.length > 1 ? `
          <div class="chart-container">
            <div class="chart-title"><i data-lucide="trending-up" class="icon"></i> Price History</div>
            ${priceChart(history)}
          </div>
        ` : ''}
        <div class="card-details" style="color:var(--muted);font-size:12px">
          <i data-lucide="eye" class="icon-sm"></i> First seen: ${new Date(a.first_seen).toLocaleDateString()} · Last seen: ${new Date(a.last_seen).toLocaleDateString()}
        </div>
      </div>
    `;

    lucide.createIcons();
  }

  function priceChart(records) {
    const W = 800, H = 180, PAD = 60;
    const prices = records.map(r => r.price);
    const minP = Math.min(...prices) - 20;
    const maxP = Math.max(...prices) + 20;
    const rangeP = maxP - minP || 1;

    const points = records.map((r, i) => {
      const x = PAD + (i / Math.max(records.length - 1, 1)) * (W - PAD * 2);
      const y = H - PAD - ((r.price - minP) / rangeP) * (H - PAD * 2);
      return { x, y, price: r.price, date: new Date(r.scraped_at) };
    });

    const line = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ');
    const dots = points.map(p => `
      <g class="chart-point">
        <circle cx="${p.x}" cy="${p.y}" r="12" fill="transparent" class="chart-hit"/>
        <circle cx="${p.x}" cy="${p.y}" r="4" class="chart-dot"/>
        <g class="chart-tooltip" transform="translate(${p.x}, ${p.y})">
          <rect x="-45" y="-38" width="90" height="28" rx="4" class="chart-tooltip-bg"/>
          <text y="-20" text-anchor="middle" class="chart-tooltip-price">$${fmt(p.price)}</text>
          <text y="20" text-anchor="middle" class="chart-tooltip-date">${p.date.toLocaleDateString()}</text>
        </g>
      </g>
    `).join('');
    const labels = [points[0], points[points.length - 1]].filter(Boolean).map(p =>
      `<text x="${p.x}" y="${H - 8}" class="chart-label" text-anchor="middle">${p.date.toLocaleDateString()}</text>`
    ).join('');
    const pLabels = [minP, maxP].map((p, i) =>
      `<text x="${PAD - 8}" y="${i === 0 ? H - PAD : PAD}" class="chart-label" text-anchor="end">$${fmt(Math.round(p))}</text>`
    ).join('');

    return `<svg viewBox="0 0 ${W} ${H}" class="chart-svg">
      <line x1="${PAD}" y1="${PAD}" x2="${PAD}" y2="${H - PAD}" class="chart-grid"/>
      <line x1="${PAD}" y1="${H - PAD}" x2="${W - PAD}" y2="${H - PAD}" class="chart-grid"/>
      <path d="${line}" class="chart-line"/>
      ${dots}${labels}${pLabels}
    </svg>`;
  }

  // Floor plans page
  async function renderFloorPlans() {
    const plans = await api('floor-plans');
    app.innerHTML = `
      <div class="fp-grid">
        ${plans.map(p => `
          <div class="fp-card">
            <div class="fp-code"><i data-lucide="layout-grid" class="icon"></i> ${p.code}</div>
            <div class="fp-specs"><i data-lucide="bed-double" class="icon-sm"></i> ${p.bedrooms} bed · <i data-lucide="bath" class="icon-sm"></i> ${p.bathrooms} bath · <i data-lucide="ruler" class="icon-sm"></i> ${fmt(p.sqft)} sqft</div>
            <div class="fp-price"><i data-lucide="shield" class="icon-sm"></i> Deposit: $${fmt(p.deposit)}</div>
            <span class="fp-badge ${p.is_renovated ? 'renovated' : 'premium'}">${p.is_renovated ? '<i data-lucide="sparkles" class="icon-sm"></i> Renovated' : '<i data-lucide="gem" class="icon-sm"></i> Premium'}</span>
          </div>
        `).join('')}
      </div>
    `;

    lucide.createIcons();
  }

  // Utilities
  function fmt(n) {
    return n ? n.toLocaleString() : '0';
  }

  function ordinal(n) {
    const s = ['th', 'st', 'nd', 'rd'];
    const v = n % 100;
    return n + (s[(v - 20) % 10] || s[v] || s[0]);
  }

  function timeAgo(date) {
    const s = Math.floor((Date.now() - date.getTime()) / 1000);
    if (s < 60) return 'just now';
    if (s < 3600) return Math.floor(s / 60) + 'm ago';
    if (s < 86400) return Math.floor(s / 3600) + 'h ago';
    return Math.floor(s / 86400) + 'd ago';
  }

  // ─── Command Palette ──────────────────────────────────────

  let paletteIdx = 0;
  let paletteItems = [];
  let paletteAllItems = [];

  function togglePalette() {
    document.querySelector('.palette-overlay') ? closePalette() : openPalette();
  }

  function openPalette() {
    const isUnit = (location.hash || '').startsWith('#/unit/');
    const cmds = [
      { icon: '🏠', title: 'Properties (Home)', action: () => { location.hash = '#/'; } },
      { icon: '📋', title: 'View Floor Plans', action: () => { location.hash = '#/floor-plans'; } },
      { icon: '↻', title: 'Re-scrape all properties', action: doScrape },
      { icon: '🌵', title: 'Browse: Desert Club', action: () => { location.hash = '#/property/desert-club'; } },
      { icon: '🏜️', title: 'Browse: Hideaway', action: () => { location.hash = '#/property/hideaway'; } },
    ];
    if (isUnit) {
      cmds.unshift({ icon: '←', title: 'Back to apartments', action: () => { history.back(); } });
    }
    cmds.push({ icon: '✕', title: 'Clear all filters', action: () => {
      filters = { property: null, beds: null, baths: null, min_price: null, max_price: null, plan: null, renovated: null, available_from: null, available_to: null, sort: 'price', order: 'asc' };
      route();
    }});

    paletteAllItems = cmds;
    paletteItems = cmds;
    paletteIdx = 0;

    const overlay = document.createElement('div');
    overlay.className = 'palette-overlay';
    overlay.innerHTML = `
      <div class="palette">
        <input class="palette-input" type="text" placeholder="Type a command…">
        <div class="palette-results">${renderPaletteItems()}</div>
      </div>`;
    document.body.appendChild(overlay);

    overlay.querySelector('.palette-overlay');
    overlay.addEventListener('click', (e) => { if (e.target === overlay) closePalette(); });

    const input = overlay.querySelector('.palette-input');
    input.addEventListener('input', () => {
      const q = input.value.toLowerCase();
      paletteItems = paletteAllItems.filter(c => !q || c.title.toLowerCase().includes(q));
      paletteIdx = 0;
      updatePaletteResults();
    });
    input.addEventListener('keydown', onPaletteKeydown);
    setTimeout(() => input.focus(), 50);
  }

  function closePalette() {
    document.querySelector('.palette-overlay')?.remove();
  }

  function onPaletteKeydown(e) {
    if (e.key === 'Escape') { closePalette(); e.preventDefault(); return; }
    if (e.key === 'ArrowDown') { e.preventDefault(); paletteIdx = Math.min(paletteIdx + 1, paletteItems.length - 1); updatePaletteResults(); }
    if (e.key === 'ArrowUp') { e.preventDefault(); paletteIdx = Math.max(paletteIdx - 1, 0); updatePaletteResults(); }
    if (e.key === 'Enter' && paletteItems[paletteIdx]) { e.preventDefault(); closePalette(); paletteItems[paletteIdx].action(); }
  }

  function updatePaletteResults() {
    const el = document.querySelector('.palette-results');
    if (el) {
      el.innerHTML = renderPaletteItems();
      const sel = el.querySelector('.palette-item.selected');
      if (sel) sel.scrollIntoView({ block: 'nearest' });
    }
  }

  function renderPaletteItems() {
    if (!paletteItems.length) return '<div class="palette-empty">No results</div>';
    return paletteItems.map((c, i) => `
      <div class="palette-item${i === paletteIdx ? ' selected' : ''}" data-palette-idx="${i}">
        <span class="palette-item-icon">${c.icon}</span>
        <span class="palette-item-title">${c.title}</span>
        <span class="palette-item-meta">Command</span>
      </div>
    `).join('');
  }

  // Palette click handler
  document.addEventListener('click', (e) => {
    const item = e.target.closest('.palette-item');
    if (item) {
      const idx = parseInt(item.dataset.paletteIdx, 10);
      if (paletteItems[idx]) { closePalette(); paletteItems[idx].action(); }
    }
  });

  // ─── Keyboard shortcuts ──────────────────────────────────

  document.addEventListener('keydown', (e) => {
    const paletteOpen = !!document.querySelector('.palette-overlay');

    // ⌘K / Ctrl+K: toggle command palette
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      togglePalette();
      return;
    }

    if (paletteOpen) return; // palette handles its own keys

    if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
    if (e.key === 'r') doScrape();
    if (e.key === '/') { e.preventDefault(); const el = $('#min-price'); if (el) el.focus(); }
    if (e.key === 'Escape') {
      filters = { property: null, beds: null, baths: null, min_price: null, max_price: null, plan: null, renovated: null, available_from: null, available_to: null, sort: 'price', order: 'asc' };
      if (location.hash === '#/' || location.hash === '' || location.hash === '#') renderDashboard();
    }
  });

  // ─── Init ────────────────────────────────────────────────

  window.addEventListener('hashchange', route);
  $('#btn-scrape').onclick = doScrape;
  $('#btn-cmd-k').onclick = togglePalette;
  loadLastScraped();
  route();
})();
