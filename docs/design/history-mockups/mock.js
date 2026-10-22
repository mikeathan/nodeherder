/*
 * Shared mockup helpers: sample navigation, app shell, synthetic history data and a small
 * SVG chart renderer with crosshair, legend toggles, rubber-band zoom and annotations.
 * Prototype code only — never imported by the application.
 */
(function (global) {
  'use strict';

  var VARIANTS = [
    { file: 'v1-history-chips.html', label: '1 · Discoverable history' },
    { file: 'v2-multi-compare.html', label: '2 · Multi-series + compare' },
    { file: 'v3-zoom-study.html', label: '3 · Zoom, layers, expand' },
    { file: 'v4-statistics.html', label: '4 · History vs statistics' },
    { file: 'v5-mobile-sheet.html', label: '5 · Mobile sheet → full' },
    { file: 'v6-mobile-nav.html', label: '6 · Bottom nav + history' }
  ];

  var NAV = [
    ['Overview', 'overview'],
    ['Home', 'home'],
    ['Devices', 'devices'],
    ['Device list', 'list'],
    ['Automations', 'automation'],
    ['History', 'chart'],
    ['Assistant', 'assistant'],
    ['Settings', 'settings']
  ];

  function el(tag, attrs, text) {
    var node = document.createElement(tag);
    Object.keys(attrs || {}).forEach(function (key) {
      if (key === 'class') node.className = attrs[key];
      else node.setAttribute(key, attrs[key]);
    });
    if (text != null) node.textContent = text;
    return node;
  }

  /* ------------------------------------------------------------ sample chrome */

  function themeState() {
    var stored = {};
    try {
      stored = JSON.parse(localStorage.getItem('nh-mock-theme') || '{}');
    } catch (error) {
      stored = {};
    }
    var state = { theme: stored.theme || 'fleece', mode: stored.mode || 'dark' };
    apply(state);
    return state;
  }

  function apply(state) {
    document.documentElement.setAttribute('data-theme', state.theme);
    document.documentElement.setAttribute('data-mode', state.mode);
    try {
      localStorage.setItem('nh-mock-theme', JSON.stringify(state));
    } catch (error) {
      /* file:// may block storage; the samples still work */
    }
  }

  function mountBar(current) {
    var state = themeState();
    var bar = el('nav', { class: 'sample-bar', 'aria-label': 'Design samples' });
    VARIANTS.forEach(function (variant) {
      var link = el('a', { href: variant.file }, variant.label);
      if (variant.file === current) link.setAttribute('aria-current', 'page');
      bar.appendChild(link);
    });
    bar.appendChild(el('span', { class: 'spacer' }));

    var themeButton = el('button', { type: 'button' }, 'Theme: ' + state.theme);
    themeButton.addEventListener('click', function () {
      state.theme = state.theme === 'fleece' ? 'daylight' : 'fleece';
      themeButton.textContent = 'Theme: ' + state.theme;
      apply(state);
    });
    var modeButton = el('button', { type: 'button' }, 'Mode: ' + state.mode);
    modeButton.addEventListener('click', function () {
      state.mode = state.mode === 'dark' ? 'light' : 'dark';
      modeButton.textContent = 'Mode: ' + state.mode;
      apply(state);
    });
    bar.appendChild(themeButton);
    bar.appendChild(modeButton);
    document.body.insertBefore(bar, document.body.firstChild);
  }

  /* ------------------------------------------------------------ app shell */

  function mountShell(options) {
    var host = document.getElementById('shell');
    if (!host) return;
    var nav = el('aside', { class: 'nav' });
    var brand = el('div', { class: 'brand' });
    brand.appendChild(el('span', { class: 'brand-dot' }));
    brand.appendChild(el('span', {}, 'NodeHerder'));
    nav.appendChild(brand);
    NAV.forEach(function (item) {
      var link = el('a', { class: 'nav-item' + (item[0] === options.active ? ' is-active' : ''), href: '#' });
      link.appendChild(el('span', { class: 'ic' }));
      link.appendChild(el('span', {}, item[0]));
      nav.appendChild(link);
    });

    var main = el('div', { class: 'main' });
    var header = el('header', { class: 'header' });
    header.appendChild(el('h1', {}, options.title));
    header.appendChild(el('span', { class: 'spacer' }));
    header.appendChild(el('span', { class: 'tiny muted' }, options.subtitle || ''));
    main.appendChild(header);
    var content = el('div', { class: 'content' + (options.wide ? ' is-wide' : ' is-narrow') });
    content.id = 'content';
    main.appendChild(content);

    var shell = el('div', { class: 'app' });
    shell.appendChild(nav);
    shell.appendChild(main);
    host.appendChild(shell);
    return {
      content: content,
      header: header
    };
  }

  /* ------------------------------------------------------------ synthetic data */

  function rng(seed) {
    var state = seed >>> 0;
    return function () {
      state = (state * 1664525 + 1013904223) >>> 0;
      return state / 4294967296;
    };
  }

  var PROFILES = {
    temperature: { base: 21.4, amp: 2.6, noise: 0.12, unit: '°C', color: 'var(--nh-kind-temp)' },
    humidity: { base: 47, amp: 9, noise: 0.5, unit: '%', color: 'var(--nh-kind-humidity)' },
    illuminance: { base: 180, amp: 150, noise: 12, unit: 'lx', color: 'var(--nh-kind-light)' },
    battery: { base: 86, amp: 1.2, noise: 0.05, unit: '%', color: 'var(--nh-ok)' },
    power: { base: 240, amp: 180, noise: 22, unit: 'W', color: 'var(--nh-kind-power)' }
  };

  function makeSeries(options) {
    var profile = PROFILES[options.kind] || PROFILES.temperature;
    var random = rng(options.seed || 7);
    var points = options.points || 140;
    var from = options.from;
    var to = options.to;
    var span = to - from;
    var data = [];
    var phase = random() * Math.PI * 2;
    for (var i = 0; i <= points; i++) {
      var ratio = i / points;
      var t = from + span * ratio;
      var daily = Math.sin(phase + ratio * Math.PI * 2);
      var drift = Math.sin(phase * 2 + ratio * Math.PI * 0.6) * 0.35;
      var value = profile.base + (options.amp == null ? profile.amp : options.amp) * daily + drift + (random() - 0.5) * profile.noise;
      data.push({ t: Math.round(t), v: Math.round(value * 10) / 10 });
    }
    return {
      id: options.id || options.kind,
      label: options.label || options.kind,
      color: options.color || profile.color,
      unit: options.unit || profile.unit,
      data: data
    };
  }

  /* ------------------------------------------------------------ formatting */

  function formatValue(value, unit) {
    if (value == null || !isFinite(value)) return '—';
    var rounded = Math.abs(value) >= 100 ? Math.round(value) : Math.round(value * 10) / 10;
    return String(rounded) + (unit ? ' ' + unit : '');
  }

  function formatStamp(t, span) {
    var date = new Date(t);
    if (span > 2 * 86400000) {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  function formatFull(t) {
    return new Date(t).toLocaleString([], {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  /* ------------------------------------------------------------ chart renderer */

  var SVG = 'http://www.w3.org/2000/svg';

  function svgEl(tag, attrs) {
    var node = document.createElementNS(SVG, tag);
    Object.keys(attrs || {}).forEach(function (key) {
      node.setAttribute(key, attrs[key]);
    });
    return node;
  }

  /**
   * chart(container, model)
   * model: { from, to, series[], unit, height, bands[], markers[], zoom, onZoom,
   *          readout (element), legend (element), yMin, yMax, area, step, bars }
   */
  function chart(container, model) {
    var state = { from: model.from, to: model.to, hidden: {}, hoverIndex: null };

    function visible() {
      return model.series.filter(function (item) {
        return !state.hidden[item.id];
      });
    }

    function bounds() {
      var min = Infinity;
      var max = -Infinity;
      visible().forEach(function (item) {
        item.data.forEach(function (point) {
          if (point.t < state.from || point.t > state.to) return;
          if (point.v < min) min = point.v;
          if (point.v > max) max = point.v;
        });
      });
      if (!isFinite(min) || !isFinite(max)) {
        min = 0;
        max = 1;
      }
      if (model.bands && model.bands.length) {
        model.bands.forEach(function (band) {
          min = Math.min(min, band.from);
          max = Math.max(max, band.to);
        });
      }
      var pad = (max - min) * 0.12 || 1;
      return { min: min - pad, max: max + pad };
    }

    function render() {
      var width = Math.max(280, container.clientWidth || 640);
      var height = model.height || 320;
      var pad = { l: 56, r: 18, t: 16, b: 30 };
      var plotW = width - pad.l - pad.r;
      var plotH = height - pad.t - pad.b;
      var y = bounds();

      function x(t) {
        return pad.l + ((t - state.from) / (state.to - state.from)) * plotW;
      }
      function yy(v) {
        return pad.t + plotH - ((v - y.min) / (y.max - y.min)) * plotH;
      }

      container.innerHTML = '';
      var svg = svgEl('svg', {
        viewBox: '0 0 ' + width + ' ' + height,
        width: width,
        height: height,
        role: 'img',
        'aria-label': (model.label || 'History chart') + ' from ' + formatStamp(state.from, state.to - state.from) + ' to ' + formatStamp(state.to, state.to - state.from)
      });

      var defs = svgEl('defs');
      var gradient = svgEl('linearGradient', { id: 'mock-area', x1: '0', y1: '0', x2: '0', y2: '1' });
      var top = svgEl('stop', { offset: '0%', 'stop-color': model.series[0].color, 'stop-opacity': '0.28' });
      var bottom = svgEl('stop', { offset: '100%', 'stop-color': model.series[0].color, 'stop-opacity': '0' });
      gradient.appendChild(top);
      gradient.appendChild(bottom);
      defs.appendChild(gradient);
      svg.appendChild(defs);

      /* threshold bands behind everything */
      (model.bands || []).forEach(function (band) {
        var bandTop = yy(Math.min(band.to, y.max));
        var bandBottom = yy(Math.max(band.from, y.min));
        svg.appendChild(
          svgEl('rect', {
            x: pad.l,
            y: bandTop,
            width: plotW,
            height: Math.max(0, bandBottom - bandTop),
            fill: band.color || 'var(--nh-accent)',
            opacity: '0.12'
          })
        );
        if (band.label) {
          var label = svgEl('text', {
            x: pad.l + 6,
            y: bandTop + 12,
            fill: 'var(--nh-text-muted)',
            'font-size': '10'
          });
          label.textContent = band.label;
          svg.appendChild(label);
        }
      });

      /* y grid + labels */
      var ticks = 4;
      for (var i = 0; i <= ticks; i++) {
        var value = y.min + ((y.max - y.min) * i) / ticks;
        var gy = yy(value);
        svg.appendChild(
          svgEl('line', {
            x1: pad.l,
            y1: gy,
            x2: width - pad.r,
            y2: gy,
            stroke: 'var(--nh-border)',
            'stroke-width': '1',
            'stroke-dasharray': i === 0 ? '' : '4 4'
          })
        );
        var tick = svgEl('text', {
          x: pad.l - 8,
          y: gy + 4,
          'text-anchor': 'end',
          fill: 'var(--nh-text-muted)',
          'font-size': '10'
        });
        tick.textContent = formatValue(value, model.unit || '');
        svg.appendChild(tick);
      }

      /* x labels */
      var span = state.to - state.from;
      for (var j = 0; j <= 4; j++) {
        var tx = pad.l + (plotW * j) / 4;
        var stamp = svgEl('text', {
          x: tx,
          y: height - 10,
          'text-anchor': j === 0 ? 'start' : j === 4 ? 'end' : 'middle',
          fill: 'var(--nh-text-muted)',
          'font-size': '10'
        });
        stamp.textContent = formatStamp(state.from + (span * j) / 4, span);
        svg.appendChild(stamp);
      }

      /* series */
      visible().forEach(function (item, index) {
        var points = item.data.filter(function (point) {
          return point.t >= state.from && point.t <= state.to;
        });
        if (!points.length) return;

        if (model.bars) {
          var barWidth = Math.max(2, plotW / Math.max(points.length, 1) - 1);
          points.forEach(function (point) {
            svg.appendChild(
              svgEl('rect', {
                x: x(point.t) - barWidth / 2,
                y: yy(point.v),
                width: barWidth,
                height: Math.max(0, pad.t + plotH - yy(point.v)),
                rx: '2',
                fill: item.color,
                opacity: '0.85'
              })
            );
          });
          return;
        }

        var d = '';
        points.forEach(function (point, index2) {
          var px = x(point.t);
          var py = yy(point.v);
          if (index2 === 0) {
            d += 'M' + px.toFixed(1) + ' ' + py.toFixed(1);
          } else if (model.step) {
            /* step-after: hold the previous value until the new sample arrives */
            d += 'L' + px.toFixed(1) + ' ' + yy(points[index2 - 1].v).toFixed(1) + 'L' + px.toFixed(1) + ' ' + py.toFixed(1);
          } else {
            d += 'L' + px.toFixed(1) + ' ' + py.toFixed(1);
          }
        });

        if (model.area !== false && index === 0 && !item.ghost) {
          var baseline = pad.t + plotH;
          svg.appendChild(
            svgEl('path', {
              d: d + 'L' + x(points[points.length - 1].t).toFixed(1) + ' ' + baseline + 'L' + x(points[0].t).toFixed(1) + ' ' + baseline + 'Z',
              fill: 'url(#mock-area)',
              stroke: 'none'
            })
          );
        }

        svg.appendChild(
          svgEl('path', {
            d: d,
            fill: 'none',
            stroke: item.color,
            'stroke-width': item.ghost ? '2' : '3',
            'stroke-dasharray': item.ghost ? '6 6' : '',
            'stroke-linecap': 'round',
            'stroke-linejoin': 'round',
            opacity: item.ghost ? '0.55' : '1'
          })
        );

        /* final sample marker keeps "latest" truthful */
        var last = points[points.length - 1];
        svg.appendChild(
          svgEl('circle', {
            cx: x(last.t),
            cy: yy(last.v),
            r: '4',
            fill: item.color,
            stroke: 'var(--nh-surface)',
            'stroke-width': '2'
          })
        );
      });

      /* automation markers */
      (model.markers || []).forEach(function (marker) {
        svg.appendChild(
          svgEl('line', {
            x1: x(marker.t),
            y1: pad.t,
            x2: x(marker.t),
            y2: pad.t + plotH,
            stroke: 'var(--nh-accent)',
            'stroke-width': '1',
            'stroke-dasharray': '3 3',
            opacity: '0.8'
          })
        );
        var markerLabel = svgEl('text', {
          x: x(marker.t) + 4,
          y: pad.t + 12,
          fill: 'var(--nh-accent)',
          'font-size': '10'
        });
        markerLabel.textContent = marker.label;
        svg.appendChild(markerLabel);
      });

      /* crosshair layer */
      var overlay = svgEl('rect', {
        x: pad.l,
        y: pad.t,
        width: plotW,
        height: plotH,
        fill: 'transparent',
        style: 'cursor:crosshair'
      });
      svg.appendChild(overlay);
      var cross = svgEl('line', {
        x1: 0,
        y1: pad.t,
        x2: 0,
        y2: pad.t + plotH,
        stroke: 'var(--nh-text-muted)',
        'stroke-width': '1',
        'stroke-dasharray': '3 3',
        opacity: '0'
      });
      svg.appendChild(cross);
      var dot = svgEl('circle', { r: '4.5', fill: model.series[0].color, stroke: 'var(--nh-surface)', 'stroke-width': '2', opacity: '0' });
      svg.appendChild(dot);
      var zoomBand = svgEl('rect', { x: pad.l, y: pad.t, width: 0, height: plotH, fill: 'var(--nh-accent)', opacity: '0.12' });
      svg.appendChild(zoomBand);
      container.appendChild(svg);

      function nearest(t) {
        var list = model.series[0].data.filter(function (point) {
          return point.t >= state.from && point.t <= state.to;
        });
        if (!list.length) return null;
        var best = list[0];
        list.forEach(function (point) {
          if (Math.abs(point.t - t) < Math.abs(best.t - t)) best = point;
        });
        return best;
      }

      function pointerT(event) {
        var box = svg.getBoundingClientRect();
        var scale = width / box.width;
        var px = (event.clientX - box.left) * scale;
        var ratio = Math.min(1, Math.max(0, (px - pad.l) / plotW));
        return state.from + ratio * (state.to - state.from);
      }

      function showCrosshair(t) {
        var point = nearest(t);
        if (!point) return;
        var px = x(point.t);
        cross.setAttribute('x1', px);
        cross.setAttribute('x2', px);
        cross.setAttribute('opacity', '1');
        dot.setAttribute('cx', px);
        dot.setAttribute('cy', yy(point.v));
        dot.setAttribute('opacity', '1');
        if (model.readout) {
          model.readout.innerHTML = '';
          model.readout.appendChild(el('span', { class: 'value' }, formatValue(point.v, model.unit)));
          model.readout.appendChild(el('span', { class: 'when' }, formatFull(point.t)));
        }
      }

      function hideCrosshair() {
        cross.setAttribute('opacity', '0');
        dot.setAttribute('opacity', '0');
        if (model.readout && model.hint) model.readout.innerHTML = '<span class="when">' + model.hint + '</span>';
      }

      var dragStart = null;
      overlay.addEventListener('pointermove', function (event) {
        if (dragStart != null) {
          var now = pointerT(event);
          var left = Math.min(x(dragStart), x(now));
          zoomBand.setAttribute('x', left);
          zoomBand.setAttribute('width', Math.abs(x(now) - x(dragStart)));
          return;
        }
        showCrosshair(pointerT(event));
      });
      overlay.addEventListener('pointerdown', function (event) {
        if (!model.zoom) return;
        dragStart = pointerT(event);
      });
      overlay.addEventListener('pointerup', function (event) {
        if (dragStart == null) return;
        var end = pointerT(event);
        var changed = Math.abs(x(end) - x(dragStart)) > 18;
        zoomBand.setAttribute('width', 0);
        var start = dragStart;
        dragStart = null;
        if (changed && model.onZoom) model.onZoom(Math.min(start, end), Math.max(start, end));
        else showCrosshair(end);
      });
      overlay.addEventListener('pointerleave', function () {
        if (dragStart == null) hideCrosshair();
      });
      overlay.addEventListener('dblclick', function () {
        if (model.onReset) model.onReset();
      });
      overlay.addEventListener(
        'wheel',
        function (event) {
          if (!model.zoom || !model.onZoom) return;
          event.preventDefault();
          var center = pointerT(event);
          var span = state.to - state.from;
          var factor = event.deltaY > 0 ? 1.25 : 0.8;
          var next = Math.min(span * 31, Math.max(5 * 60000, span * factor));
          var ratio = (center - state.from) / span;
          model.onZoom(center - next * ratio, center + next * (1 - ratio));
        },
        { passive: false }
      );

      if (model.legend) renderLegend();
      hideCrosshair();
      return { width: width, height: height };
    }

    function renderLegend() {
      var host = model.legend;
      host.innerHTML = '';
      model.series.forEach(function (item) {
        var pressed = String(!state.hidden[item.id]);
        var button = el('button', { type: 'button', 'aria-pressed': pressed });
        button.style.setProperty('--c', item.color);
        button.appendChild(el('span', { class: 'line' }));
        button.appendChild(el('span', {}, item.label));
        button.addEventListener('click', function () {
          state.hidden[item.id] = !state.hidden[item.id];
          button.setAttribute('aria-pressed', String(!state.hidden[item.id]));
          render();
        });
        host.appendChild(button);
      });
    }

    /* paint immediately; the observer only re-renders after a real size change */
    render();
    var observer = new ResizeObserver(function () {
      render();
    });
    observer.observe(container);

    return {
      render: render,
      setRange: function (from, to) {
        state.from = from;
        state.to = to;
        render();
      },
      range: function () {
        return { from: state.from, to: state.to };
      },
      destroy: function () {
        observer.disconnect();
        container.innerHTML = '';
      }
    };
  }

  /* ------------------------------------------------------------ chip helper */

  function chips(host, options) {
    var active = options.active;
    function paint() {
      host.innerHTML = '';
      options.items.forEach(function (item) {
        var button = el('button', { type: 'button', class: 'chip' + (item.value === active ? ' is-active' : '') });
        if (item.color) {
          var swatch = el('span', { class: 'swatch' });
          swatch.style.setProperty('--swatch', item.color);
          button.appendChild(swatch);
        }
        button.appendChild(el('span', {}, item.label));
        if (item.count != null) button.appendChild(el('span', { class: 'count' }, item.count));
        button.addEventListener('click', function () {
          active = item.value;
          paint();
          options.onChange(item.value);
        });
        host.appendChild(button);
      });
    }
    paint();
    return {
      value: function () {
        return active;
      }
    };
  }

  global.NH = {
    VARIANTS: VARIANTS,
    mountBar: mountBar,
    mountShell: mountShell,
    makeSeries: makeSeries,
    chart: chart,
    chips: chips,
    formatValue: formatValue,
    formatStamp: formatStamp,
    formatFull: formatFull,
    el: el
  };
})(window);
