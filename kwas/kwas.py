"""Kmer-wide association study."""
import argparse
from datetime import datetime
from typing import Any, Iterable, Tuple

import numpy as np
import pandas as pd
import statsmodels.api as sm
from setproctitle import setproctitle

from hasloader.hasloader import HasLoader


def timers():
    c = {a * 10**b for a in [1, 2, 5] for b in range(10)}
    i = 0
    t = datetime.now()

    def inc():
        nonlocal c, i, t
        i += 1
        if i in c:
            d = datetime.now() - t
            print(f'\r{d} ({d/i}) {i}', end='')

    def done():
        nonlocal t, i
        d = datetime.now() - t
        if i == 0:
            print(f'\r{d}')
        else:
            print(f'\r{d} ({d/i}) {i}')

    return inc, done


def regression(x, y, k):
    """Runs regression on x and y and returns the result in a dict."""
    y_binary = len(set(y)) == 2

    if y_binary:
        sm_result = sm.Logit(y, x).fit(method='nm')
    else:
        sm_result = sm.OLS(y, x).fit()

    result = {'key': k, 'n': int(x['kmer'].sum())}
    coef_interval = sm_result.conf_int()

    result[
        'rsquared'] = sm_result.prsquared if y_binary else sm_result.rsquared
    for col in x.columns:
        result[f'{col}_coef'] = sm_result.params[col]
        result[f'{col}_pval'] = sm_result.pvalues[col]
        result[f'{col}_coef_025'] = coef_interval.loc[col, 0]
        result[f'{col}_coef_975'] = coef_interval.loc[col, 1]

    return result


def xy_iter_gen(exe_fname, has_fname, cov_fname,
                y_col) -> Iterable[Tuple[Any, Any, Any]]:
    """Returns an iterator of (X,Y,key) tuples."""
    print('This run:', has_fname)

    print('Loading covariate matrix')
    if cov_fname.endswith('.h5'):
        covariates_df = pd.read_hdf(cov_fname)
    elif cov_fname.endswith('.csv'):
        covariates_df = pd.read_csv(cov_fname)
    elif cov_fname.endswith('.tsv'):
        covariates_df = pd.read_csv(cov_fname, sep='\t')
    else:
        raise TypeError(
            f'unsupported file type for {cov_fname}: want .csv or .tsv or .h5')
    covariates_df['const'] = 1
    print('Covariates:', covariates_df.columns)
    good_rows = ~covariates_df.isna().any(axis=1)
    assert len(good_rows) == len(covariates_df)

    ld = HasLoader(exe_fname, has_fname)
    n = -1
    has = np.array([0] * len(covariates_df))

    inc, done = timers()

    for kmer in ld:
        n += 1
        has[:] = 0
        has[kmer['samples']] = 1
        covariates_df['kmer'] = has
        # TODO(amit): This can be optimized to avoid [good_rows], which takes ~17%
        #   of the run time.
        yield (
            covariates_df[[x for x in covariates_df.columns
                           if x != y_col]][good_rows],
            covariates_df[[y_col]][good_rows],
            kmer['kmer'],
        )
        inc()
    done()


def run(exe_fname: str, has_fname: str, out_fname: str, cov_fname: str,
        y_col: str):
    """Runs the KWAS process."""
    xy_iter = xy_iter_gen(exe_fname, has_fname, cov_fname, y_col)
    dicts = (regression(x, y, k) for x, y, k in xy_iter)
    df = pd.DataFrame(dicts)

    print('Writing to:', out_fname)
    df.to_csv(out_fname, index=False)
    print('Done')


def main():
    setproctitle('kwas')

    arg_parser = argparse.ArgumentParser()
    arg_parser.add_argument('-i', required=True, help="Input HAS file")
    arg_parser.add_argument('-c',
                            required=True,
                            help="Input covariates CSV/TSV/H5 file")
    arg_parser.add_argument('-o', required=True, help="Output CSV file")
    arg_parser.add_argument('-x', required=True, help="Hastojson executable")
    arg_parser.add_argument('-y', required=True, help="Y column name")
    args: argparse.Namespace = arg_parser.parse_args()

    run(args.x, args.i, args.o, args.c, args.y)


if __name__ == '__main__':
    main()
