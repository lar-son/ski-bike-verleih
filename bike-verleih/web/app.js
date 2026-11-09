// Fahrrad-Verleih Frontend: reines JS, spricht nur mit der Go-API unter /api.
const AUSWEIS_TYPEN = ['Personalausweis', 'Reisepass', 'Führerschein', 'Aufenthaltstitel'];
const KATEGORIEN = ['City', 'Trekking', 'Mountainbike', 'E-Bike', 'Kinder'];

// ---------- Helfer ----------
const $ = (sel, el = document) => el.querySelector(sel);
const esc = s => String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const eur = cent => new Intl.NumberFormat('de-AT', { style: 'currency', currency: 'EUR' }).format(cent / 100);
const fdt = iso => new Date(iso).toLocaleString('de-AT', { weekday: 'short', day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
const toCent = v => Math.round(parseFloat(String(v).replace(',', '.')) * 100) || 0;
const fromCent = c => (c / 100).toFixed(2);
const pad = n => String(n).padStart(2, '0');
const localDate = d => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
const badge = status => `<span class="badge b-${esc(status)}">${esc(status)}</span>`;

async function api(method, path, body) {
  const opt = { method, headers: {} };
  if (body instanceof FormData) opt.body = body;
  else if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
  const res = await fetch('/api' + path, opt);
  const data = res.status === 204 ? null : await res.json().catch(() => null);
  if (!res.ok) throw new Error(data?.error || `Fehler ${res.status}`);
  return data;
}

let toastTimer;
function toast(msg, isError = false) {
  const t = $('#toast');
  t.textContent = msg;
  t.className = 'show' + (isError ? ' error' : '');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (t.className = ''), isError ? 5000 : 2800);
}

// Führt eine Aktion aus, zeigt Fehler als Toast; true bei Erfolg.
async function attempt(fn, okMsg) {
  try { await fn(); if (okMsg) toast(okMsg); return true; }
  catch (e) { toast(e.message, true); return false; }
}

// Dialog mit Formular. onSubmit(formData) wirft bei Fehler -> Dialog bleibt offen.
function openDialog(title, fieldsHtml, onSubmit, submitLabel = 'Speichern') {
  const dlg = $('#dlg'), form = $('#dlgForm');
  form.innerHTML = `<h3>${esc(title)}</h3>${fieldsHtml}
    <div class="form-actions">
      <button type="button" class="btn grey" id="dlgCancel">Abbrechen</button>
      <button class="btn green">${esc(submitLabel)}</button>
    </div>`;
  $('#dlgCancel').onclick = () => dlg.close();
  form.onsubmit = async e => {
    e.preventDefault();
    if (await attempt(() => onSubmit(new FormData(form)))) dlg.close();
  };
  dlg.showModal();
}

function select(name, options, selected, extra = '') {
  return `<select name="${name}" ${extra}>${options.map(o => `<option ${o === selected ? 'selected' : ''}>${esc(o)}</option>`).join('')}</select>`;
}

// ---------- Routing ----------
const routes = { '': dashboard, bike: bikeView, personen: personenView, reservierungen: reservierungenView, ausleihe: ausleiheView };

async function render() {
  const route = location.hash.replace(/^#\/?/, '');
  document.querySelectorAll('#nav a').forEach(a => a.classList.toggle('active', a.dataset.route === route));
  const app = $('#app');
  try { await (routes[route] || dashboard)(app); }
  catch (e) { app.innerHTML = `<h1>Fehler</h1><p class="hint warn">${esc(e.message)}</p>`; }
}
window.addEventListener('hashchange', render);
window.addEventListener('DOMContentLoaded', render);

// ---------- Dashboard (SPEC R1–R4) ----------
async function dashboard(app) {
  const [st, rentals] = await Promise.all([api('GET', '/stats'), api('GET', '/rentals')]);
  const active = rentals.filter(r => r.status === 'ausgeliehen').sort((a, b) => new Date(a.end) - new Date(b.end));
  app.innerHTML = `
    <h1>Dashboard</h1>
    <div class="stats">
      <div class="stat s1"><div class="label">Verfügbare Fahrräder</div><div class="value">${st.verfuegbar}</div><div class="sub">${st.reserviert} reserviert</div></div>
      <div class="stat s2"><div class="label">Ausgeliehene Fahrräder</div><div class="value">${st.ausgeliehen}</div><div class="sub">${st.ueberfaellig ? st.ueberfaellig + ' überfällig' : ''}</div></div>
      <div class="stat s3"><div class="label">Gesamt Fahrräder</div><div class="value">${st.gesamt}</div><div class="sub"></div></div>
      <div class="stat s4"><div class="label">Registrierte Personen</div><div class="value">${st.personen}</div><div class="sub"></div></div>
    </div>

    <h2>Schnellzugriff</h2>
    <div class="actions">
      <a class="btn" href="#/bike">Fahrräder verwalten</a>
      <a class="btn" href="#/personen">Personen verwalten</a>
      <a class="btn green" href="#/ausleihe">Neue Ausleihe</a>
    </div>

    <h2>Aktuell ausgeliehen</h2>
    ${active.length ? `<div class="tablewrap"><table>
      <tr><th>Fahrrad</th><th>Person</th><th>Seit</th><th>Rückgabe bis</th><th></th></tr>
      ${active.map(r => `<tr>
        <td>${esc(r.bikeName)}</td><td>${esc(r.personName)}</td><td>${fdt(r.start)}</td>
        <td>${fdt(r.end)} ${r.overdue ? '<span class="badge b-ueberfaellig">überfällig</span>' : ''}</td>
        <td><button class="btn small green" data-return="${r.id}">Zurücknehmen</button></td></tr>`).join('')}
    </table></div>` : '<div class="empty">Aktuell ist kein Fahrrad ausgeliehen.</div>'}`;
  app.querySelectorAll('[data-return]').forEach(b => b.onclick = async () => {
    if (await attempt(() => api('POST', `/rentals/${b.dataset.return}/return`), 'Fahrrad zurückgenommen')) render();
  });
}

// ---------- Fahrrad-Verwaltung (SPEC R5, R10) ----------
async function bikeView(app) {
  const bikes = await api('GET', '/bikes');
  app.innerHTML = `
    <h1>Fahrrad-Verwaltung</h1>
    <div class="toolbar"><span class="hint">${bikes.length} Fahrräder im Bestand</span><button class="btn green" id="add">+ Neues Fahrrad</button></div>
    ${bikes.length ? `<div class="tablewrap"><table>
      <tr><th>Nr.</th><th>Marke</th><th>Modell</th><th class="num">Rahmen</th><th>Kategorie</th><th class="num">Preis / Std.</th><th class="num">Preis / Tag</th><th>Status</th><th></th></tr>
      ${bikes.map(s => `<tr>
        <td>${s.id}</td><td>${esc(s.marke)}</td><td>${esc(s.modell)}</td><td class="num">${s.rahmenCm} cm</td><td>${esc(s.kategorie)}</td>
        <td class="num">${eur(s.preisStundeCent)}</td><td class="num">${s.preisTagCent ? eur(s.preisTagCent) : '–'}</td>
        <td>${s.ausgeliehen ? badge('ausgeliehen') : '<span class="badge b-verfügbar">verfügbar</span>'}</td>
        <td class="row-actions"><button class="btn small" data-edit="${s.id}">Bearbeiten</button><button class="btn small red" data-del="${s.id}">Löschen</button></td></tr>`).join('')}
    </table></div>` : '<div class="empty">Noch keine Fahrräder. Lege den ersten an.</div>'}`;
  $('#add').onclick = () => bikeForm(null);
  app.querySelectorAll('[data-edit]').forEach(b => b.onclick = () => bikeForm(bikes.find(s => s.id == b.dataset.edit)));
  app.querySelectorAll('[data-del]').forEach(b => b.onclick = async () => {
    const s = bikes.find(x => x.id == b.dataset.del);
    if (confirm(`${s.marke} ${s.modell} wirklich löschen?`) && await attempt(() => api('DELETE', `/bikes/${s.id}`), 'Fahrrad gelöscht')) render();
  });
}

function bikeForm(s) {
  s = s || { marke: '', modell: '', rahmenCm: 54, kategorie: 'City', preisStundeCent: 400, preisTagCent: 1800 };
  openDialog(s.id ? 'Fahrrad bearbeiten' : 'Neues Fahrrad', `
    <div class="row"><div><label>Marke</label><input name="marke" required value="${esc(s.marke)}"></div>
    <div><label>Modell</label><input name="modell" required value="${esc(s.modell)}"></div></div>
    <div class="row"><div><label>Rahmengröße (cm)</label><input name="rahmenCm" type="number" min="30" max="70" required value="${s.rahmenCm}"></div>
    <div><label>Kategorie</label>${select('kategorie', KATEGORIEN, s.kategorie)}</div></div>
    <div class="row"><div><label>Preis pro Stunde (€)</label><input name="preisStunde" type="number" min="0.5" step="0.5" required value="${fromCent(s.preisStundeCent)}"></div>
    <div><label>Preis pro Tag (€, 24 h)</label><input name="preisTag" type="number" min="0" step="0.5" value="${s.preisTagCent ? fromCent(s.preisTagCent) : ''}"></div></div>
    <div class="hint">Der Tagespreis ist optional. Mehrere Stunden kosten nie mehr als der Tagespreis. Geänderte Preise gelten nur für neue Ausleihen.</div>`,
    async fd => {
      const body = {
        marke: fd.get('marke'), modell: fd.get('modell'), rahmenCm: +fd.get('rahmenCm'), kategorie: fd.get('kategorie'),
        preisStundeCent: toCent(fd.get('preisStunde')), preisTagCent: fd.get('preisTag') ? toCent(fd.get('preisTag')) : 0,
      };
      await (s.id ? api('PUT', `/bikes/${s.id}`, body) : api('POST', '/bikes', body));
      toast('Fahrrad gespeichert');
      render();
    });
}

// ---------- Personen (SPEC R6–R8) ----------
async function personenView(app) {
  const persons = await api('GET', '/persons');
  app.innerHTML = `
    <h1>Personen</h1>
    <div class="toolbar"><span class="hint">${persons.length} registrierte Personen</span><button class="btn green" id="add">+ Neue Person</button></div>
    ${persons.length ? `<div class="tablewrap"><table>
      <tr><th>Nr.</th><th>Name</th><th>E-Mail</th><th>Telefon</th><th>Ausweis</th><th></th></tr>
      ${persons.map(p => `<tr>
        <td>${p.id}</td><td>${esc(p.vorname)} ${esc(p.nachname)}</td><td>${esc(p.email)}</td><td>${esc(p.telefon)}</td>
        <td>${ausweisCell(p)}</td>
        <td class="row-actions"><button class="btn small" data-edit="${p.id}">Bearbeiten</button><button class="btn small red" data-del="${p.id}">Löschen</button></td></tr>`).join('')}
    </table></div>` : '<div class="empty">Noch keine Personen.</div>'}`;
  $('#add').onclick = () => personForm(null);
  app.querySelectorAll('[data-edit]').forEach(b => b.onclick = () => personForm(persons.find(p => p.id == b.dataset.edit)));
  app.querySelectorAll('[data-del]').forEach(b => b.onclick = async () => {
    const p = persons.find(x => x.id == b.dataset.del);
    if (confirm(`${p.vorname} ${p.nachname} wirklich löschen?`) && await attempt(() => api('DELETE', `/persons/${p.id}`), 'Person gelöscht')) render();
  });
}

function ausweisCell(p) {
  const parts = [];
  if (p.ausweisTyp) parts.push(`${esc(p.ausweisTyp)} ${esc(p.ausweisNr)}`);
  if (p.dokument) parts.push(`<a href="/api/persons/${p.id}/document" target="_blank" rel="noopener">📎 ${esc(p.dokument.fileName)}</a>`);
  return parts.length ? parts.join(' · ') : '<span class="badge b-warn">fehlt</span>';
}

function personForm(p) {
  p = p || { vorname: '', nachname: '', email: '', telefon: '', ausweisTyp: '', ausweisNr: '', dokument: null };
  openDialog(p.id ? 'Person bearbeiten' : 'Neue Person', `
    <div class="row"><div><label>Vorname</label><input name="vorname" required value="${esc(p.vorname)}"></div>
    <div><label>Nachname</label><input name="nachname" required value="${esc(p.nachname)}"></div></div>
    <div class="row"><div><label>E-Mail</label><input name="email" type="email" value="${esc(p.email)}"></div>
    <div><label>Telefon</label><input name="telefon" value="${esc(p.telefon)}"></div></div>
    <h2 style="margin:22px 0 0;font-size:16px">Ausweis</h2>
    <div class="hint">Für eine Ausleihe: Art und Nummer eingeben <b>oder</b> eine Datei hochladen.</div>
    <div class="row"><div><label>Art</label>
      <select name="ausweisTyp"><option value="">– keine Angabe –</option>${AUSWEIS_TYPEN.map(t => `<option ${t === p.ausweisTyp ? 'selected' : ''}>${t}</option>`).join('')}</select></div>
    <div><label>Nummer</label><input name="ausweisNr" value="${esc(p.ausweisNr)}"></div></div>
    <label>Ausweis hochladen (PNG, JPEG oder PDF, max. 5 MB)</label>
    <input type="file" name="file" accept="image/png,image/jpeg,application/pdf">
    ${p.dokument ? `<div class="docinfo">Aktuell: <a href="/api/persons/${p.id}/document" target="_blank" rel="noopener">📎 ${esc(p.dokument.fileName)}</a>
      <label><input type="checkbox" name="removeDoc"> Dokument entfernen</label></div>` : ''}`,
    async fd => {
      const body = Object.fromEntries(['vorname', 'nachname', 'email', 'telefon', 'ausweisTyp', 'ausweisNr'].map(k => [k, fd.get(k)]));
      const saved = await (p.id ? api('PUT', `/persons/${p.id}`, body) : api('POST', '/persons', body));
      const file = fd.get('file');
      if (file && file.size > 0) {
        const up = new FormData(); up.append('file', file);
        await api('POST', `/persons/${saved.id}/document`, up);
      } else if (fd.get('removeDoc')) {
        await api('DELETE', `/persons/${saved.id}/document`);
      }
      toast('Person gespeichert');
      render();
    });
}

// ---------- Reservierungen (SPEC R9) ----------
let rentalFilter = 'alle';
async function reservierungenView(app) {
  const all = await api('GET', '/rentals');
  const list = rentalFilter === 'alle' ? all : all.filter(r => r.status === rentalFilter);
  app.innerHTML = `
    <h1>Reservierungen</h1>
    <div class="toolbar">
      <div class="right"><label style="margin:0">Status</label>
        <select id="filter" style="width:auto">${['alle', 'reserviert', 'ausgeliehen', 'zurückgegeben', 'storniert'].map(s => `<option ${s === rentalFilter ? 'selected' : ''}>${s}</option>`).join('')}</select></div>
      <a class="btn green" href="#/ausleihe">+ Neue Ausleihe</a>
    </div>
    ${list.length ? `<div class="tablewrap"><table>
      <tr><th>Nr.</th><th>Fahrrad</th><th>Person</th><th>Beginn</th><th class="num">Dauer</th><th>Ende</th><th class="num">Preis</th><th>Status</th><th></th></tr>
      ${list.map(r => `<tr>
        <td>${r.id}</td><td>${esc(r.bikeName)}</td><td>${esc(r.personName)}</td><td>${fdt(r.start)}</td>
        <td class="num">${r.hours} h</td><td>${fdt(r.end)}</td><td class="num">${eur(r.priceCent)}</td>
        <td>${badge(r.status)} ${r.overdue ? '<span class="badge b-ueberfaellig">überfällig</span>' : ''}</td>
        <td class="row-actions">${rentalButtons(r)}</td></tr>`).join('')}
    </table></div>` : '<div class="empty">Keine Einträge.</div>'}`;
  $('#filter').onchange = e => { rentalFilter = e.target.value; render(); };
  app.querySelectorAll('[data-act]').forEach(b => b.onclick = async () => {
    const msg = { pickup: 'Fahrrad ausgegeben', return: 'Fahrrad zurückgenommen', cancel: 'Reservierung storniert' }[b.dataset.act];
    if (await attempt(() => api('POST', `/rentals/${b.dataset.id}/${b.dataset.act}`), msg)) render();
  });
}

function rentalButtons(r) {
  const btn = (act, label, cls) => `<button class="btn small ${cls}" data-act="${act}" data-id="${r.id}">${label}</button>`;
  if (r.status === 'reserviert') return btn('pickup', 'Ausgeben', 'green') + btn('cancel', 'Stornieren', 'red');
  if (r.status === 'ausgeliehen') return btn('return', 'Zurücknehmen', 'green');
  return '';
}

// ---------- Neue Ausleihe (SPEC R8, R12, R13, R15) ----------
async function ausleiheView(app) {
  const persons = await api('GET', '/persons');
  const now = new Date();
  const hourOpts = Array.from({ length: 24 }, (_, h) => `<option value="${h}" ${h === now.getHours() ? 'selected' : ''}>${pad(h)}:00 Uhr</option>`).join('');
  app.innerHTML = `
    <h1>Neue Ausleihe</h1>
    <form id="rentalForm">
      <label>Person</label>
      <select id="person" required>
        <option value="">– bitte wählen –</option>
        ${persons.map(p => `<option value="${p.id}" ${(p.ausweisTyp && p.ausweisNr) || p.dokument ? '' : 'disabled'}>${esc(p.vorname)} ${esc(p.nachname)}${(p.ausweisTyp && p.ausweisNr) || p.dokument ? '' : ' – Ausweis fehlt'}</option>`).join('')}
      </select>
      <div class="hint" id="personHint">Personen ohne Ausweis sind gesperrt. <a href="#/personen">Ausweis unter „Personen“ ergänzen</a>.</div>

      <div class="row">
        <div><label>Beginn – Datum</label><input type="date" id="date" min="${localDate(now)}" value="${localDate(now)}" required></div>
        <div><label>Beginn – Uhrzeit (volle Stunde)</label><select id="hour">${hourOpts}</select></div>
      </div>

      <label>Dauer (Stunden)</label>
      <input type="number" id="hours" min="1" max="168" value="4" required>
      <div class="chips" id="chips">${[1, 2, 3, 4, 8, 24, 48].map(h => `<button type="button" class="chip" data-h="${h}">${h < 24 ? h + ' h' : h / 24 + (h === 24 ? ' Tag' : ' Tage')}</button>`).join('')}</div>

      <label>Fahrrad (nur im Zeitraum verfügbare)</label>
      <select id="bike" required></select>
      <div class="hint" id="bikeHint"></div>

      <div class="price-box"><span>Ende: <b id="endTxt">–</b></span><span>Gesamtpreis <span class="big" id="priceTxt">–</span></span></div>
      <div class="form-actions"><a class="btn grey" href="#/">Abbrechen</a><button class="btn green" id="submit" disabled>Ausleihe starten</button></div>
    </form>`;

  const el = id => $('#' + id, app);
  let offers = [];

  const startDate = () => new Date(`${el('date').value}T${pad(el('hour').value)}:00:00`);

  async function refresh() {
    const hours = +el('hours').value;
    document.querySelectorAll('#chips .chip').forEach(c => c.classList.toggle('on', +c.dataset.h === hours));
    const start = startDate();
    const hourStart = new Date(); hourStart.setMinutes(0, 0, 0);
    el('submit').textContent = start > new Date() ? 'Reservieren' : 'Ausleihe starten';
    if (!(hours >= 1 && hours <= 168) || isNaN(start)) { setOffers([], 'Dauer: 1 bis 168 Stunden.', true); return; }
    if (start < hourStart) { setOffers([], 'Der Beginn liegt in der Vergangenheit.', true); return; }
    el('endTxt').textContent = fdt(new Date(start.getTime() + hours * 3600e3));
    try {
      offers = await api('GET', `/bikes/available?start=${encodeURIComponent(start.toISOString())}&hours=${hours}`);
    } catch (e) { setOffers([], e.message, true); return; }
    setOffers(offers, offers.length ? `${offers.length} Fahrräder verfügbar.` : 'Kein Fahrrad in diesem Zeitraum verfügbar – anderen Zeitraum wählen.', !offers.length);
  }

  function setOffers(list, hint, warn) {
    const bikeSel = el('bike'), prev = bikeSel.value;
    offers = list;
    bikeSel.innerHTML = list.length
      ? list.map(o => `<option value="${o.id}">${esc(o.marke)} ${esc(o.modell)} · ${o.rahmenCm} cm · ${esc(o.kategorie)} – ${eur(o.priceCent)}</option>`).join('')
      : '<option value="">– keiner verfügbar –</option>';
    if (list.some(o => o.id == prev)) bikeSel.value = prev;
    el('bikeHint').textContent = hint;
    el('bikeHint').className = 'hint' + (warn ? ' warn' : '');
    updateState();
  }

  function updateState() {
    const o = offers.find(x => x.id == el('bike').value);
    el('priceTxt').textContent = o ? eur(o.priceCent) : '–';
    el('submit').disabled = !(o && el('person').value);
  }

  ['date', 'hour', 'hours'].forEach(id => el(id).addEventListener('input', refresh));
  el('bike').addEventListener('change', updateState);
  el('person').addEventListener('change', updateState);
  document.querySelectorAll('#chips .chip').forEach(c => c.onclick = () => { el('hours').value = c.dataset.h; refresh(); });

  el('rentalForm').onsubmit = async e => {
    e.preventDefault();
    const body = { bikeId: +el('bike').value, personId: +el('person').value, start: startDate().toISOString(), hours: +el('hours').value };
    if (await attempt(() => api('POST', '/rentals', body), 'Ausleihe gespeichert')) location.hash = '#/reservierungen';
    else refresh(); // Verfügbarkeit hat sich evtl. geändert
  };

  refresh();
}
