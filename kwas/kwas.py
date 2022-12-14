"""Kmer-wide association study (pronounced "kvass")."""
import argparse
from typing import Any, Iterable, Tuple

import numpy as np
import pandas as pd
import statsmodels.api as sm
from hasloader.hasloader import HasLoader
from setproctitle import setproctitle


def regression(x, y, k):
    """Runs regression on x and y and returns the result in a dict."""
    y_binary = len(set(y)) == 2

    if y_binary:
        sm_result = sm.Logit(y, x).fit(method='nm')
    else:
        sm_result = sm.OLS(y, x).fit()

    result = {'key': k, 'n': int(x['has'].sum())}
    coef_interval = sm_result.conf_int()

    result[
        'rsquared'] = sm_result.prsquared if y_binary else sm_result.rsquared
    for col in x.columns:
        result[f'{col}_coef'] = sm_result.params[col]
        result[f'{col}_pval'] = sm_result.pvalues[col]
        result[f'{col}_coef_025'] = coef_interval.loc[col, 0]
        result[f'{col}_coef_975'] = coef_interval.loc[col, 1]

    return result


def xy_iter_gen(exe_fname, has_fname,
                cov_fname) -> Iterable[Tuple[Any, Any, Any]]:
    """Returns an iterator of (X,Y,key) tuples."""
    print('This run:', has_fname)

    print('Loading covariate matrix')
    covariates_df = pd.read_hdf(cov_fname)
    print('Covariates:', covariates_df.columns)
    good_rows = ~covariates_df.isna().any(axis=1)
    assert len(good_rows) == len(covariates_df)

    ld = HasLoader(exe_fname, has_fname)
    n = -1
    has = np.array([0] * len(covariates_df))

    for kmer in ld:
        n += 1
        has[:] = 0
        has[kmer['samples']] = 1
        covariates_df['has'] = has
        # TODO(amit): This can be optimized to avoid [good_rows], which takes ~17%
        #   of the run time.
        yield (
            covariates_df[[x for x in covariates_df.columns
                           if x != 'bmi']][good_rows],
            covariates_df[['bmi']][good_rows],
            kmer['kmer'],
        )


def run(exe_fname: str, has_fname: str, out_fname: str, cov_fname: str):
    """Runs the KWAS process."""
    xy_iter = xy_iter_gen(exe_fname, has_fname, cov_fname)
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
                            help="Input covariates H5 file")
    arg_parser.add_argument('-o', required=True, help="Output CSV file")
    arg_parser.add_argument('-x', required=True, help="Hastojson executable")
    args: argparse.Namespace = arg_parser.parse_args()

    run(args.x, args.i, args.o, args.c)


if __name__ == '__main__':
    main()
