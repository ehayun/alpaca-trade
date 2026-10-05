// Price chart: daily candles (default) or close-price line, from /api/bars.
import { api, showMessage, hideMessage } from './common.js';

const LWC = window.LightweightCharts;
const form = document.getElementById('bars-form');
const msg = document.getElementById('bars-msg');
const title = document.getElementById('chart-title');
const box = document.getElementById('chart');
const toggles = document.querySelectorAll('[name=chart-type]');

const dark = matchMedia('(prefers-color-scheme: dark)').matches;
const chart = LWC.createChart(box, {
  autoSize: true,
  layout: {
    background: { color: 'transparent' },
    textColor: dark ? '#dee2e6' : '#212529',
    fontFamily: 'Heebo, system-ui, sans-serif',
  },
  grid: {
    vertLines: { color: dark ? '#343a40' : '#e9ecef' },
    horzLines: { color: dark ? '#343a40' : '#e9ecef' },
  },
  localization: { locale: 'he-IL', dateFormat: 'dd/MM/yyyy' },
  timeScale: { borderVisible: false },
  rightPriceScale: { borderVisible: false },
});

let candles = [];
let series = null;

function draw() {
  if (series) chart.removeSeries(series);
  const type = document.querySelector('[name=chart-type]:checked').value;
  if (type === 'line') {
    series = chart.addSeries(LWC.LineSeries, { color: '#2a78d6', lineWidth: 2 });
    series.setData(candles.map((c) => ({ time: c.time, value: c.close })));
  } else if (type === 'bars') {
    series = chart.addSeries(LWC.BarSeries, { upColor: '#198754', downColor: '#dc3545' });
    series.setData(candles);
  } else {
    series = chart.addSeries(LWC.CandlestickSeries, {
      upColor: '#198754', downColor: '#dc3545', borderVisible: false,
      wickUpColor: '#198754', wickDownColor: '#dc3545',
    });
    series.setData(candles);
  }
  chart.timeScale().fitContent();
}

async function load(symbol, days) {
  hideMessage(msg);
  form.querySelector('button').disabled = true;
  try {
    const data = await api(`/api/bars?symbol=${encodeURIComponent(symbol)}&days=${encodeURIComponent(days)}`);
    candles = data.candles;
    const sym = document.createElement('bdi');
    sym.textContent = data.symbol;
    title.replaceChildren(sym, ` · ${days} ימים · ${candles.length} ימי מסחר`);
    draw();
    document.dispatchEvent(new CustomEvent('symbol-loaded', { detail: data.symbol }));
  } catch (err) {
    showMessage(msg, err.message);
  } finally {
    form.querySelector('button').disabled = false;
  }
}

form.addEventListener('submit', (e) => {
  e.preventDefault();
  const symbol = form.symbol.value.trim().toUpperCase();
  history.replaceState(null, '', `?symbol=${encodeURIComponent(symbol)}&days=${encodeURIComponent(form.days.value)}`);
  load(symbol, form.days.value);
});
toggles.forEach((t) => t.addEventListener('change', () => candles.length && draw()));

// Shareable links: ?symbol=AAPL&days=90&chart=bars loads the chart on open.
const params = new URLSearchParams(location.search);
const typeInput = document.querySelector(`[name=chart-type][value="${CSS.escape(params.get('chart') || '')}"]`);
if (typeInput) typeInput.checked = true;
if (params.get('symbol')) {
  form.symbol.value = params.get('symbol').toUpperCase();
  if (params.get('days')) form.days.value = params.get('days');
  load(form.symbol.value, form.days.value);
}
