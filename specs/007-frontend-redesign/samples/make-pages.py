#!/usr/bin/env python3
"""Generate the per-design HTML pages (source form). Run: python3 make-pages.py"""
import json, os
HERE = os.path.dirname(os.path.abspath(__file__))
D = {
 'hearth': dict(fonts='family=Figtree:wght@400;500;600;700;800', fx='dot grid', layout=False,
   presets=[('daylight','Daylight',['#f4f2ee','#0f8bd6']),('teak','Teak',['#f3ece1','#b8611f']),('avocado','Avocado',['#eef0e2','#5e7d1e'])],
   accents=['#0f8bd6','#d9480f','#2b8a3e','#7048e8','#c2255c','#0c8599']),
 'floorplan': dict(fonts='family=Barlow:wght@400;500;600;700&family=Barlow+Condensed:wght@500;600;700', fx='graph paper', layout=True,
   presets=[('blueprint','Blueprint',['#f5f8fc','#1d2b45']),('sepia','Sepia',['#f7f4ee','#3b2f22']),('mint','Mint',['#e4f1ec','#163a33'])],
   accents=['#1f4f99','#136b5a','#a1430f','#6a3fb5','#b0245c','#1d2b45']),
 'workbench': dict(fonts='family=IBM+Plex+Sans:wght@400;500;600&family=IBM+Plex+Mono:wght@400;500', fx='row stripes', layout=True,
   presets=[('graphite','Graphite',['#22262d','#5aa9e6']),('solar','Solar',['#fdf6e3','#268bd2']),('forest','Forest',['#1e2622','#7cc68a'])],
   accents=['#5aa9e6','#e0a33b','#7cc68a','#c792ea','#f07178','#89ddff']),
 'brief': dict(fonts='family=Atkinson+Hyperlegible:ital,wght@0,400;0,700;1,400', fx='marker highlights', layout=True,
   presets=[('paper','Paper',['#fafbfc','#0f7e6e']),('slate','Slate',['#eef1f5','#3a55b4']),('night','Night',['#1a1f29','#ffe79a'])],
   accents=['#0f7e6e','#3a55b4','#9a3f9e','#b4482e','#1f2430','#6b7d1b']),
 'deck': dict(fonts='family=Lexend:wght@300;400;500;600;700', fx='glass + glow', layout=True,
   presets=[('lagoon','Lagoon',['#13252c','#ffb547']),('ember','Ember',['#24170f','#ff8a3d']),('frost','Frost',['#e6eef0','#1d6fa3'])],
   accents=['#ffb547','#6fd3e8','#ff7a8a','#9be37a','#c9a6ff','#ffffff']),
}
for k, v in D.items():
    extra = {'presets': [{'id': i, 'name': n, 'swatch': s} for i, n, s in v['presets']], 'accents': v['accents'], 'effectsLabel': v['fx']}
    layout = '<script src="designs/%s.js"></script>\n' % k if v['layout'] else ''
    name = k.capitalize()
    html = f'''<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>NodeHerder · {name} design sample</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?{v['fonts']}&display=swap">
<link rel="stylesheet" href="shared/base.css">
<link rel="stylesheet" href="designs/{k}.css">
</head>
<body>
<div id="app"><p style="padding:2rem;font-family:system-ui">Loading NodeHerder sample… (JavaScript required)</p></div>
<script src="shared/designs.js"></script>
<script>window.NH_DESIGN = NH_DESIGN_FOR('{k}', {json.dumps(extra)});</script>
<script src="shared/icons.js"></script>
<script src="shared/data.js"></script>
{layout}<script src="shared/app.js"></script>
</body>
</html>
'''
    open(os.path.join(HERE, k + '.html'), 'w').write(html)
    print('wrote', k + '.html')
