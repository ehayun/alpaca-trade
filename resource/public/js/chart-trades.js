// Trades on the price chart: fill markers, the position entry line and open order lines.
// A click on the chart (with קנייה / מכירה picked) puts that price into the order form as a limit order.
import { api, num } from './common.js';

const LWC = window.LightweightCharts;
const BUY = '#0072B2'; // the colors of up / down candles
const SELL = '#E69F00';
const LABEL = { buy: 'קנייה', sell: 'מכירה' };

let series = null; // the price series; chart.js replaces it on redraw
let markers = null; // markers plugin of that series
let lines = []; // price lines on that series
let times = []; // candle days, sorted; a marker must sit on one
let weekly = false;
let symbol = '';
let days = 0;
let fills = [];
let positions = [];
let orders = [];
let positionColor = '#495057';

// candleDay is the candle a fill day sits on: the same day, or on a weekly chart the week's candle
// (the last candle on or before that day). Undefined when there is none.
function candleDay(day) {
  if (!weekly) return times.includes(day) ? day : undefined;
  return times.findLast((t) => t <= day);
}

function render() {
  if (!series) return;
  markers.setMarkers(fills.map((f) => ({ ...f, time: candleDay(f.day) })).filter((f) => f.time).map((f) => ({
    time: f.time,
    position: f.side === 'buy' ? 'belowBar' : 'aboveBar',
    shape: f.side === 'buy' ? 'arrowUp' : 'arrowDown',
    color: f.side === 'buy' ? BUY : SELL,
    text: `${f.side === 'buy' ? '+' : '−'}${num(f.qty)} @ ${Number(f.price).toFixed(2)}`,
  })));
  for (const l of lines) series.removePriceLine(l);
  lines = [];
  const pos = positions.find((p) => p.symbol === symbol);
  if (pos) {
    lines.push(series.createPriceLine({
      price: Number(pos.avgEntryPrice), color: positionColor, lineWidth: 2,
      lineStyle: LWC.LineStyle.Solid, title: `פוזיציה ${num(pos.qty)}`,
    }));
  }
  for (const o of orders) {
    if (o.symbol !== symbol || !o.limitPrice) continue; // market orders have no price to draw
    lines.push(series.createPriceLine({
      price: Number(o.limitPrice), color: o.side === 'buy' ? BUY : SELL, lineWidth: 1,
      lineStyle: LWC.LineStyle.Dashed, title: `${LABEL[o.side] || o.side} ${num(o.qty)}`,
    }));
  }
}

async function loadFills(days) {
  const want = symbol;
  try {
    const data = await api(`/api/fills?symbol=${encodeURIComponent(want)}&days=${encodeURIComponent(days)}`);
    if (want === symbol) fills = data.fills;
  } catch {
    if (want === symbol) fills = []; // the chart still works without markers; trade errors show in the side panel
  }
  render();
}

// initTrades hooks the chart click once; dark sets the position line color.
export function initTrades(chart, dark) {
  positionColor = dark ? '#dee2e6' : '#495057';
  chart.subscribeClick((p) => {
    const side = document.querySelector('[name=chart-click]:checked')?.value;
    if (!side || !series || !p.point || p.paneIndex !== 0) return; // not the MACD pane
    const price = series.coordinateToPrice(p.point.y);
    if (price === null || price <= 0) return;
    const f = document.getElementById('order-form');
    f.side.value = side;
    f.type.value = 'limit';
    f.type.dispatchEvent(new Event('change')); // shows the limit price field
    f.limitPrice.value = price.toFixed(2);
    document.getElementById('tab-order').click();
    f.qty.focus();
  });
}

// setTradesSeries is called after each redraw with the new price series, its candles and their time frame.
export function setTradesSeries(s, candles, wk) {
  if (s !== series) {
    series = s;
    markers = LWC.createSeriesMarkers(s, []);
    lines = []; // the old lines went away with the old series
  }
  times = candles.map((c) => c.time);
  weekly = wk;
  render();
}

// loadTrades fetches the fills of a new chart symbol / day range.
export function loadTrades(sym, d) {
  symbol = sym;
  days = d;
  fills = [];
  loadFills(d);
}

// trade.js sends this after each refresh (page load, new order, cancel); a market order may have filled.
document.addEventListener('trades-changed', (e) => {
  ({ positions, orders } = e.detail);
  if (symbol) loadFills(days);
  else render();
});
