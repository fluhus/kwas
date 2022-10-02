"""Analyzes the results of smfqhg.py. Looks at species distribution."""
import gzip
import json
import math
from collections import Counter

import myfisher
import numpy as np
import pandas as pd
from matplotlib import pyplot as plt

DIR = 'TODO'
PATHWAY_FILE = f'{DIR}/genes_pathway.json.gz'
KO_FILE = f'{DIR}/genes_ko.json.gz'
KO_DEFS_FILE = f'{DIR}/ko_def.json.gz'
KO_CAT_FILE = f'{DIR}/ko_cat.json.gz'


def cut_suffix(a, prefix):
    return [prefix if x.startswith(prefix + '_') else x for x in a]


def onto_others(a, thr=0.05):
    b = a[a >= thr]
    b['Other'] = a[a < thr].sum()
    return b


def onto_others_dict(a, thr=0.05):
    keys = sorted(a, key=lambda x: -a[x])
    b = {k: a[k] for k in keys if a[k] >= thr}
    b['Other'] = sum(filter(lambda x: x < thr, a.values()))
    return b


def pie_params(a):
    cmap = plt.get_cmap('Pastel1')
    return {
        'autopct': '%.0f%%',
        'colors': [cmap(i) for i in range(len(a))],
        'startangle': 90,
        'shadow': True,
        'explode': [0.03] * len(a),
        'counterclock': False,
    }


def by_species():
    print('by_species2')
    tax = pd.read_csv('/tmp/amitmit/taxonomy.csv.gz')
    d = json.load(gzip.open('/tmp/amitmit/fishers.bwt.json.gz'))
    sig_species = set(d['sig'])
    found_species = set(d['found'])
    del d

    print(tax.columns)

    tax['Phylum'] = cut_suffix(tax['Phylum'].values, 'Firmicutes')
    tax['Class'] = cut_suffix(tax['Class'].values, 'Clostridia')
    taxfound = tax[tax['SGB'].isin(found_species)]
    taxsig = tax[tax['SGB'].isin(sig_species)]

    for col in ('Phylum', 'Class', 'Order'):
        tvcf = taxfound[col].value_counts() / len(taxfound)
        tvcs = taxsig[col].value_counts() / len(taxsig)

        plt.subplot(121)
        plt.pie(oth := onto_others(tvcf, 0.03),
                labels=oth.index,
                **pie_params(oth))
        plt.xlabel(f'Found species ({len(taxfound)})')
        plt.subplot(122)
        plt.pie(oth := onto_others(tvcs, 0.03),
                labels=oth.index,
                **pie_params(oth))
        plt.xlabel(f'Enriched species ({len(taxsig)})')
        plt.tight_layout()
        plt.show()


def by_kegg():
    print('Loading tests')
    d = json.load(gzip.open('/tmp/amitmit/fishers.dmnd.json.gz'))
    enrch = frozenset(d['sig'])
    found = frozenset(d['found'])
    nfound, nenrch = len(found), len(enrch)
    print('Found', nfound, 'enriched', nenrch)
    assert len(enrch - found) == 0, f'{enrch - found}'
    del d

    print('Loading KO')
    ko = json.load(gzip.open(KO_FILE))

    print('Counting')
    cnt_found = Counter(k for x in found for k in ko.get(x, []))
    cnt_enrch = Counter(k for x in enrch for k in ko.get(x, []))
    nopwy = [x for x in enrch if x not in ko]

    del ko, found, enrch
    print(len(cnt_found), len(cnt_enrch))
    print(f'Not found {len(nopwy)}/{nenrch} ({100*len(nopwy)/nenrch:.0f}%)')

    fishes = []
    print('Fishing')
    for pwy in cnt_enrch:
        sp = cnt_enrch[pwy]
        nsp = cnt_found[pwy] - sp
        snp = nenrch - sp
        nsnp = nfound - nenrch - nsp
        f = myfisher.fisher(sp, nsp, snp, nsnp, 'greater')
        fishes.append((pwy, *f))
    myfisher.reset()

    fishes = [{
        'ko': x[0],
        'odr': -1 if math.isinf(x[1]) else x[1],
        'sig': x[2] <= 0.05 / len(fishes),
    } for x in fishes]
    fishes = sorted(fishes, key=lambda x: (not x['sig'], -x['odr']))

    print('Tests made:', len(fishes))
    print('Enriched:', sum(x['sig'] for x in fishes))
    print(fishes[:3])
    json.dump(
        {
            'fisher': fishes,
            'noko': nopwy
        },
        gzip.open('/tmp/amitmit/ko.json.gz', 'wt'),
        indent=2,
    )


def restore_inf(x):
    return math.inf if x == -1 else x


def post_kegg():
    data = json.load(gzip.open('/tmp/amitmit/ko.json.gz'))
    data = data['fisher']
    for x in data:
        x['odr'] = restore_inf(x['odr'])
    sig = (x for x in data if x['sig'])

    cats = json.load(gzip.open(KO_CAT_FILE, 'rt'))
    print(len(cats), 'categories')
    unneeded_cats = {
        'Brite Hierarchies',
        'Not Included in Pathway or Brite',
    }

    kos = (x['ko'][3:] for x in data)
    found_counts = Counter(cat for ko in kos for cat in cats[ko]
                           if cat not in unneeded_cats)
    kos = (x['ko'][3:] for x in sig)
    sig_counts = Counter(cat for ko in kos for cat in cats[ko]
                         if cat not in unneeded_cats)

    found_draw = found_counts.most_common(10)
    sig_draw = [(x[0], sig_counts[x[0]]) for x in found_draw]

    plt.style.use('ggplot')

    plt.subplot(121).set_title('All found KOs')
    plt.barh(
        list(range(len(found_draw))),
        [x[1] for x in found_draw],
        tick_label=[x[0] for x in found_draw],
    )
    plt.gca().invert_yaxis()
    plt.subplot(122).set_title('Enriched KOs')
    plt.barh(
        list(range(len(sig_draw))),
        [x[1] for x in sig_draw],
    )
    # Remove tick numbers from righthand bars.
    ticks = np.arange(len(sig_draw))
    plt.yticks(ticks, [''] * len(ticks))
    plt.tight_layout()
    plt.gca().invert_yaxis()
    plt.show()


by_species()
by_kegg()
post_kegg()
