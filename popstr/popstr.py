import json
import os
from collections import defaultdict

import numpy as np
import pandas as pd
from matplotlib import pyplot as plt
from numpy.linalg import norm
from setproctitle import setproctitle
from sklearn.decomposition import PCA
from sklearn.manifold import MDS
from sklearn.preprocessing import StandardScaler

setproctitle('popstr')
plt.style.use('ggplot')

# Fill these.
INPUT_JSON = '/tmp/amitmit/has.json'
OUTPUT_FILE = '/tmp/amitmit/comps.json'
TMP_DIR = '/tmp/amitmit'


def read_has_json(n=None) -> pd.DataFrame:
    """Reads kmers from hastojson output."""
    h5_fname = f'{TMP_DIR}/table.h5'
    if os.path.exists(h5_fname):
        print('Loading h5')
        return pd.read_hdf(h5_fname)
    raw = (json.loads(line) for line in open(INPUT_JSON))
    dicts = ({
        'kmer': line['kmer'],
        **{x: 1
           for x in line['samples']}
    } for line in raw)
    print('Building dataframe')
    df = pd.DataFrame(dicts).set_index('kmer').transpose()
    print('Fixing NAs')
    df.replace(np.NaN, 0, inplace=True)
    print('Converting type')
    df = df.astype('int8')
    print('Saving h5')
    df.to_hdf(h5_fname, 'default')
    return df


def random_rows_cols(a, r, c):
    """Returns indexes of random selection of r rows and c columns."""
    rows = subset(a.shape[0], r)
    cols = subset(a.shape[1], c)
    return rows, cols


def shuffle_rows_cols(mm):
    """Returns mm with its rows and columns shuffled."""
    rows, cols = random_rows_cols(mm, mm.shape[0], mm.shape[1])
    return mm[rows][:, cols]


def mini_pca(a, rows, cols, b=None, n=2):
    """Projects a (or b) of a PCA space created with the given row and columns
    indexes."""
    if b is None:
        b = a
    return PCA(n).fit(a[rows][:, cols]).transform(b[:, cols])


def my_mds(x, n=2):
    """Runs MDS on the given matrix."""
    return MDS(n).fit_transform(x, init=PCA(n).fit_transform(x))


def cosine(a, b):
    """Returns the cosine similarity between a and b."""
    return np.dot(a, b) / (norm(a) * norm(b))


def my_mds_cosine(x, n=2):
    """Runs MDS using cosine dissimilarity."""
    dists = np.array([[1 - cosine(a, b) for a in x] for b in x])
    return MDS(n, dissimilarity='precomputed').fit_transform(
        dists,
        init=PCA(n).fit_transform(x),
    )


def subset(a, b):
    """Returns a random subset of a of size b."""
    return np.random.choice(a, b, replace=False)


def pca_distances(rows, cols, drows, n=2):
    """Returns an array of pairwise distances of elements in m after PCA."""
    mini = mini_pca(m, rows, cols, b=m2, n=n)
    mini = StandardScaler().fit_transform(mini)
    mini = mini[drows]
    d = np.array(
        [norm(u - v) for i, u in enumerate(mini) for v in mini[i + 1:]])
    d /= norm(d) * 2**0.5
    return d


def groupby(items, key_func, value_func):
    """"Groups items into lists by their key function."""
    result = defaultdict(list)
    for x in items:
        result[key_func(x)].append(value_func(x))
    return result


def create_final_pca(n):
    """Creates the final projection matrix for use in downstream analysis."""
    m = read_has_json()
    header = list(m.columns)
    m = m.values
    print('Shape:', m.shape)
    print('PCA')
    pca = PCA(n).fit(m)
    del m

    comp = pca.components_
    evr = pca.explained_variance_ratio_
    print('Explained variance:', evr, 'Sum:', evr.sum())
    del pca

    print('Writing to ' + OUTPUT_FILE)
    with open(OUTPUT_FILE, 'wt') as f:
        for row in comp:
            json.dump({h: v for h, v in zip(header, row)}, f)
            f.write('\n')

    evr_file = OUTPUT_FILE[:-5] + '.explnvrnc.json'
    print('Writing explained variance to:', evr_file)
    json.dump(evr.tolist(), open(evr_file, 'wt'))


if True:
    create_final_pca(10)
    exit()

m = read_has_json()
m = m.values
print('Shuffling')
m = shuffle_rows_cols(m)
print(m.shape, m.dtype)
rr, cc = m.shape

m2 = m[rr // 2:]
m = m[:rr // 2]
rr, cc = m.shape


def plot_subsample_projections():
    for i, a in enumerate([1000, 100, 30, 10, 3, 1]):
        plt.subplot(231 + i)
        print('PCA')
        randr, randc = random_rows_cols(m, rr // a, cc // a)
        mini = mini_pca(m, randr, randc, b=m2)
        plt.scatter(mini[:, 0], mini[:, 1], alpha=0.3)
        plt.xlabel(f'{rr//a} samples, {cc//a} k-mers')
    plt.subplot(232)
    plt.title('Population Structure PCA for\nDifferent Data Subsamples')
    plt.tight_layout()
    plt.show()


plot_subsample_projections()


def plot_distances_mds(rats):
    max_iter = 3
    dists = []
    rrows = subset(rr, rr // 10)
    groups = []
    labels = []
    for rat in rats:
        for a in range(min(rat, max_iter)):
            for b in range(min(rat, max_iter)):
                rrange = np.arange(round(rr / rat * a),
                                   round(rr / rat * (a + 1)))
                crange = np.arange(round(cc / rat * b),
                                   round(cc / rat * (b + 1)))
                print(f'PCA {rat} {a} {b}')
                dists.append(pca_distances(rrange, crange, rrows, n=2))
                groups.append(f'{rr//rat} samples, {cc//rat} k-mers')
                labels.append(f'{rat},{a},{b}')
    dists = np.array(dists)
    mds = my_mds_cosine(dists)
    for k, v in groupby(zip(mds, groups), lambda x: x[1],
                        lambda x: x[0]).items():
        arr = np.array(v)
        plt.scatter(arr[:, 0], arr[:, 1], alpha=0.5, label=k)
    plt.legend()
    plt.title('PCoA of Distance Vectors for\nDifferent Subsample Sizes')
    plt.tight_layout()
    plt.show()


plot_distances_mds([1000, 100, 10, 3, 1])
