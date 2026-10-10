// Price chart: daily candles (default), bars or close-price line, with EMA 20/100/200,
// and the slow stochastic (14, 3, 3) in a pane below, from /api/bars.
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
let series = null; // the price series; the only one replaced on redraw

const STOCH_PANE = 1;
const STOCH_HEIGHT = 140; // px

// Color-blind safe (Okabe-Ito): blue up, orange down. Up candles are hollow as a non-color cue.
const UP = '#0072B2';
const DOWN = '#E69F00';

// EMA lines differ in color and dash, so they stay apart without relying on color.
const EMA_STYLE = {
  20: { color: '#CC79A7', lineStyle: LWC.LineStyle.Solid },
  100: { color: '#009E73', lineStyle: LWC.LineStyle.Dashed },
  200: { color: dark ? '#dee2e6' : '#495057', lineStyle: LWC.LineStyle.Dotted },
};
// Stochastic: %K solid, %D dashed, in colors not used by the price pane.
const STOCH_STYLE = {
  k: { color: '#56B4E9', lineStyle: LWC.LineStyle.Solid, label: '%K (14, 3)' },
  d: { color: '#D55E00', lineStyle: LWC.LineStyle.Dashed, label: '%D (3)' },
};
const CSS_DASH = { [LWC.LineStyle.Solid]: 'solid', [LWC.LineStyle.Dashed]: 'dashed', [LWC.LineStyle.Dotted]: 'dotted' };

const legend = document.getElementById('chart-legend');
function legendItem(text, s) {
  const item = document.createElement('span');
  item.className = 'ema-key';
  item.style.setProperty('--ema-color', s.color);
  item.style.setProperty('--ema-dash', CSS_DASH[s.lineStyle]);
  item.textContent = text;
  legend.append(item);
}
for (const [period, s] of Object.entries(EMA_STYLE)) legendItem(`EMA ${period}`, s);
for (const s of Object.values(STOCH_STYLE)) legendItem(s.label, s);

// Indicator lines are created once and only get new data. Lightweight Charts deletes a pane when its
// last series is removed (the stochastic pane would then move up into the price pane), so the EMA
// lines keep the price pane alive while the price series is swapped.
const LINE = { lineWidth: 2, priceLineVisible: false, lastValueVisible: false, crosshairMarkerVisible: false };
const emaSeries = Object.fromEntries(
  Object.entries(EMA_STYLE).map(([period, style]) => [period, chart.addSeries(LWC.LineSeries, { ...style, ...LINE })]),
);
// The stochastic pane always shows 0..100, with dotted lines at 80 (overbought) and 20 (oversold).
const stochSeries = Object.fromEntries(['k', 'd'].map((key) => {
  const { label, ...style } = STOCH_STYLE[key];
  const opts = { ...style, ...LINE, autoscaleInfoProvider: () => ({ priceRange: { minValue: 0, maxValue: 100 } }) };
  return [key, chart.addSeries(LWC.LineSeries, opts, STOCH_PANE)];
}));
for (const level of [80, 20]) {
  stochSeries.k.createPriceLine({
    price: level, color: dark ? '#6c757d' : '#adb5bd', lineWidth: 1, lineStyle: LWC.LineStyle.Dotted, axisLabelVisible: true,
  });
}
stochSeries.k.priceScale().applyOptions({ scaleMargins: { top: 0.08, bottom: 0.08 } }); // default top margin shows a "120" tick
chart.panes()[STOCH_PANE].setHeight(STOCH_HEIGHT);

function setIndicators(data) {
  for (const e of data.emas) emaSeries[e.period]?.setData(e.points);
  stochSeries.k.setData(data.stoch.k);
  stochSeries.d.setData(data.stoch.d);
}

function draw() {
  if (series) chart.removeSeries(series);
  const type = document.querySelector('[name=chart-type]:checked').value;
  if (type === 'line') {
    series = chart.addSeries(LWC.LineSeries, { color: '#2a78d6', lineWidth: 2 });
    series.setData(candles.map((c) => ({ time: c.time, value: c.close })));
  } else if (type === 'bars') {
    series = chart.addSeries(LWC.BarSeries, { upColor: UP, downColor: DOWN });
    series.setData(candles);
  } else {
    series = chart.addSeries(LWC.CandlestickSeries, {
      upColor: 'transparent', downColor: DOWN,
      borderUpColor: UP, borderDownColor: DOWN,
      wickUpColor: UP, wickDownColor: DOWN,
    });
    series.setData(candles);
  }
  // Show all candles plus an empty margin after the last one (5% of the range, at least 2 bars).
  const n = candles.length;
  chart.timeScale().setVisibleLogicalRange({ from: -0.5, to: n - 0.5 + Math.max(2, Math.round(n * 0.05)) });
}

async function load(symbol, days) {
  hideMessage(msg);
  form.querySelector('button').disabled = true;
  try {
    const data = await api(`/api/bars?symbol=${encodeURIComponent(symbol)}&days=${encodeURIComponent(days)}`);
    candles = data.candles;
    setIndicators(data);
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
