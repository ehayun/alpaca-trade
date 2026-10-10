// EMA scanner side panel: start a scan, show progress, list the matches with live prices;
// a click loads the chart and info. On open, a saved scan (if any) is offered instead of a new scan.
import { api, showMessage, hideMessage, num, usd, dateTime } from './common.js';

const $ = (id) => document.getElementById(id);
const btn = $('scan-btn');
const statusEl = $('scan-status');
const msg = $('scan-msg');
const list = $('scan-list');
const form = $('bars-form');
const saved = $('scan-saved');

const POLL_MS = 2000;
const stacked = matchMedia('(max-width: 991.98px)'); // Bootstrap lg breakpoint
let shown = ''; // symbol on the chart, highlighted in the list

function statusText(s) {
  if (s.running) return `סורק… ${num(s.done)} מתוך ${num(s.total)} מניות`;
  if (!s.finishedAt) return 'עדיין לא נסרק.';
  return `נסרקו ${num(s.scanned)} מניות · ${dateTime(s.finishedAt)} · נמצאו ${num(s.matches.length)} (עד ${num(s.maxSpreadPct)}%)`;
}

function item(m) {
  const b = document.createElement('button');
  b.type = 'button';
  b.className = 'list-group-item list-group-item-action d-flex align-items-center gap-2';
  b.dataset.symbol = m.symbol;
  b.title = `EMA 20: ${m.ema20.toFixed(2)} · EMA 100: ${m.ema100.toFixed(2)} · EMA 200: ${m.ema200.toFixed(2)}`;
  const sym = document.createElement('bdi');
  sym.className = 'fw-medium';
  sym.textContent = m.symbol;
  const price = document.createElement('bdi');
  price.className = 'small scan-price';
  price.textContent = usd(m.close); // close at scan time until the live price arrives
  const spread = document.createElement('bdi');
  spread.className = 'small text-body-secondary ms-auto';
  spread.textContent = `${m.spreadPct.toFixed(2)}%`;
  b.append(sym, price, spread);
  b.addEventListener('click', () => {
    form.symbol.value = m.symbol;
    form.requestSubmit();
    // When the panel is stacked above the chart (narrow screens), bring the chart into view.
    if (stacked.matches) form.scrollIntoView({ behavior: 'smooth' });
  });
  return b;
}

function highlight() {
  for (const b of list.children) {
    const on = b.dataset.symbol === shown;
    b.classList.toggle('active', on);
    if (on) b.setAttribute('aria-current', 'true');
    else b.removeAttribute('aria-current');
  }
}

// Live prices for the listed symbols, in one request. On failure the scan-time close stays.
async function refreshPrices(matches) {
  if (!matches.length) return;
  try {
    const { prices } = await api(`/api/prices?symbols=${encodeURIComponent(matches.map((m) => m.symbol).join(','))}`);
    for (const b of list.children) {
      const p = prices[b.dataset.symbol];
      if (p !== undefined) b.querySelector('.scan-price').textContent = usd(p);
    }
  } catch (err) {
    console.warn('live prices:', err.message);
  }
}

function render(s) {
  btn.disabled = s.running;
  statusEl.textContent = statusText(s);
  if (s.failed) showMessage(msg, 'הסריקה נכשלה, נסו שוב');
  else hideMessage(msg);
  if (!s.running) {
    list.replaceChildren(...s.matches.map(item));
    highlight();
    refreshPrices(s.matches);
  }
}

// Poll until the scan finishes; the server keeps scanning even if the page is closed.
async function follow(s) {
  render(s);
  while (s.running) {
    await new Promise((r) => setTimeout(r, POLL_MS));
    try {
      s = await api('/api/scan');
    } catch (err) {
      showMessage(msg, err.message);
      btn.disabled = false;
      return;
    }
    render(s);
  }
}

async function startScan() {
  saved.hidden = true;
  btn.disabled = true;
  try {
    follow(await api('/api/scan', { method: 'POST' }));
  } catch (err) {
    showMessage(msg, err.message);
    btn.disabled = false;
  }
}

btn.addEventListener('click', startScan);
$('scan-saved-rescan').addEventListener('click', startScan);

document.addEventListener('symbol-loaded', (e) => {
  shown = e.detail;
  highlight();
});

// On open: follow a running scan; otherwise offer the saved scan, if there is one.
async function init() {
  const s = await api('/api/scan');
  if (s.running) return follow(s);
  const last = await api('/api/scan/saved');
  if (!last) return render(s);
  $('scan-saved-text').textContent =
    `נמצאה סריקה שמורה מ-${dateTime(last.finishedAt)} (${num(last.matches.length)} מניות, עד ${num(last.maxSpreadPct)}%).`;
  $('scan-saved-use').onclick = () => {
    saved.hidden = true;
    render(last);
  };
  statusEl.textContent = '';
  saved.hidden = false;
}

init().catch((err) => showMessage(msg, err.message));
