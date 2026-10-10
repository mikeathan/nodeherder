#!/usr/bin/env python3
"""Generate the per-design HTML pages (source form). Run: python3 make-pages.py"""
import json, os
HERE = os.path.dirname(os.path.abspath(__file__))
THEMES = [('daylight','Daylight',['#f4f2ee','#0f8bd6']),('teak','Teak',['#f3ece1','#b8611f']),('avocado','Avocado',['#eef0e2','#5e7d1e']),
  ('lagoon','Lagoon',['#13252c','#6fd3e8']),('ember','Ember',['#24170f','#ff8a3d']),('frost','Frost',['#182433','#9cc9ff']),
  ('fleece','Fleece',['#eef6fa','#1b7fb0']),('fjord','Fjord',['#0f1a20','#4fc3cf']),('plum','Plum',['#f6f1f5','#8e3b7f']),('graphite','Graphite',['#1c1c1e','#e8e8e8'])]
D = {
 'hearth': dict(fonts='family=Figtree:wght@400;500;600;700;800', fx='dot grid', layout=False, css=['hearth'], logo=False,
   presets=THEMES, accents=['#0f8bd6','#d9480f','#2b8a3e','#7048e8','#c2255c','#0c8599']),
 'deck': dict(fonts='family=Lexend:wght@300;400;500;600;700', fx='glass + glow', layout=True, css=['deck'], logo=False,
   presets=THEMES, accents=['#ffb547','#6fd3e8','#ff7a8a','#9be37a','#c9a6ff','#ffffff']),
 'hearthpanel': dict(fonts='family=Figtree:wght@400;500;600;700;800&family=Lexend:wght@300;400;500', fx='dot grid', layout=True, css=['hearth', 'hearthpanel'], logo=True,
   presets=THEMES, accents=['#1b7fb0','#0f8bd6','#d9480f','#2b8a3e','#8e3b7f','#0e7c86']),
}
for k, v in D.items():
    extra = {'presets': [{'id': i, 'name': n, 'swatch': s} for i, n, s in v['presets']], 'accents': v['accents'], 'effectsLabel': v['fx'], 'useLogo': v['logo']}
    layout = '<script src="designs/%s.js"></script>\n' % k if v['layout'] else ''
    name = {'hearthpanel': 'Hearth + Panel'}.get(k, k.capitalize())
    css = '\n'.join('<link rel="stylesheet" href="designs/%s.css">' % x for x in v['css'])
    logo = '<script src="shared/logo.js"></script>\n' if v['logo'] else ''
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
{css}
<link rel="stylesheet" href="shared/themes.css">
</head>
<body>
<div id="app"><p style="padding:2rem;font-family:system-ui">Loading NodeHerder sample… (JavaScript required)</p></div>
<script src="shared/designs.js"></script>
<script>window.NH_DESIGN = NH_DESIGN_FOR('{k}', {json.dumps(extra)});</script>
{logo}<script src="shared/icons.js"></script>
<script src="shared/data.js"></script>
{layout}<script src="shared/app.js"></script>
</body>
</html>
'''
    open(os.path.join(HERE, k + '.html'), 'w').write(html)
    print('wrote', k + '.html')
