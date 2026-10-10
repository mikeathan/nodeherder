#!/usr/bin/env python3
"""Inline each design page (CSS + JS) into one self-contained HTML file in standalone/.
Run from anywhere: python3 build-standalone.py"""
import os, re

HERE = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(HERE, 'standalone')
os.makedirs(OUT, exist_ok=True)

def read(rel):
    with open(os.path.join(HERE, rel), encoding='utf-8') as f:
        return f.read()

def inline(html):
    html = re.sub(r'<link rel="stylesheet" href="((?:shared|designs)/[^"]+)">',
                  lambda m: '<style>\n' + read(m.group(1)) + '\n</style>', html)
    html = re.sub(r'<script src="((?:shared|designs)/[^"]+)"></script>',
                  lambda m: '<script>\n' + read(m.group(1)).replace('</script', '<\\/script') + '\n</script>', html)
    return html

for name in ['index', 'hearth', 'floorplan', 'workbench', 'brief', 'deck']:
    html = inline(read(name + '.html'))
    if name != 'index':
        html = html.replace("NH_DESIGN_FOR('", "NH_DESIGN_FOR_ALL('")
        html = html.replace('<script>window.NH_DESIGN', "<script>window.NH_DESIGN_FOR_ALL = (id, x) => NH_DESIGN_FOR(id, Object.assign({ defaultRoute: '#/all' }, x));</script>\n<script>window.NH_DESIGN", 1)
    if name == 'index':
        html = html.replace('href="../', 'href="../../')  # docs links one level deeper
    with open(os.path.join(OUT, name + '.html'), 'w', encoding='utf-8') as f:
        f.write(html)
    print('wrote standalone/%s.html (%d KB)' % (name, len(html) // 1024))
