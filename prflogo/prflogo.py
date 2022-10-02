"""Creates logos for kmer profiles."""
import json
import sys

import logomaker
import pandas as pd
from matplotlib import pyplot as plt
from scipy.stats import entropy

# Fit labels for showing the logos together, rather than each individually.
BATCH_LABELS = True

maps = (json.loads(line) for line in sys.stdin)
dfs = (pd.DataFrame(m).set_index('pos') for m in maps)

for i, df in enumerate(dfs):
    df = df[df['n'] > 0]
    ent = sum([entropy(x, base=2) for x in df[['A', 'C', 'G', 'T']].values])
    ent /= len(df) - 33

    if BATCH_LABELS and i == 10:
        plt.figure(figsize=(15, 3.2), dpi=300)
    else:
        plt.figure(figsize=(15, 3), dpi=300)

    ax1: plt.Axes = plt.subplot(211)
    logo = logomaker.Logo(df[['A', 'C', 'G', 'T']], ax=ax1)
    # Trim bar bounds to compensate for its excess width.
    trim = 0.5
    ax1.plot([101 + trim, 133 - trim], [0, 0], 'k', linewidth=12)
    if not BATCH_LABELS:
        ax1.set_xlabel('position (black stripe = kmer)')
    ax1.set_ylabel('2 - entropy')
    # ax1.set_title(f'Entropy: {ent:.3f}')
    ax1.set_title(f'Entropy percentile {i*10} (avg. entropy = {ent:.3f})')

    ax2: plt.Axes = plt.subplot(212)
    plt.bar(df.index.values, df['n'].values)
    ax2.set_xlim(ax1.get_xlim())
    ax2.set_ylabel('# observations')

    if BATCH_LABELS and i == 10:
        ax2.set_xlabel('position (black stripe = kmer)')

    plt.tight_layout()
    plt.show()
