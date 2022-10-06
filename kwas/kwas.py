"""Kmer-wide association study (pronounced "kvass")."""
import argparse
from typing import Any, Iterable, Tuple

import pandas as pd
import statsmodels.api as sm
from setproctitle import setproctitle

from loader import KmerHasLoader


def regression(x, y, k):
    """Runs regression on x and y and returns the result in a dict."""
    y_binary = len(set(y)) == 2

    if y_binary:
        sm_result = sm.Logit(y, x).fit(
            method='nm')  # TODO: understand why nm is the best method
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


def keep_common(df1, df2):
    """Returns copies of df1 and df2 containing only the rows with common
    keys, in the same order."""
    common = set(df1.index.values) & set(df2.index.values)
    c1 = [x for x in df1.index.values if x in common]
    c2 = [x for x in df2.index.values if x in common]
    return df1.loc[c1], df2.loc[c2]


def xy_iter_gen(samples_fname, has_fname,
                cov_fname) -> Iterable[Tuple[Any, Any, Any]]:
    """Returns an iterator of (X,Y,key) tuples."""
    print('This run:', samples_fname, has_fname)

    print('Loading covariate matrix')
    covariates_df = pd.read_hdf(cov_fname)
    print('Covariates:', covariates_df.columns)

    print('Creating loader')
    ld = KmerHasLoader(samples_fname, has_fname)

    n = -1
    for kmer_df in ld:
        n += 1
        if 'has' not in covariates_df.columns:
            covariates_df, _ = keep_common(covariates_df, kmer_df)
            unique_idx = {x: i for i, x in enumerate(kmer_df.index.values)}
            unique_idx = [unique_idx[x] for x in covariates_df.index.values]

        covariates_df['has'] = kmer_df.iloc[unique_idx].values
        yield covariates_df[[x for x in covariates_df.columns if x != 'bmi'
                             ]], covariates_df[['bmi']], kmer_df.attrs['kmer']


def run(samples_fname: str, has_fname: str, out_fname: str, cov_fname: str):
    """Runs the KWAS process."""
    xy_iter = xy_iter_gen(samples_fname, has_fname, cov_fname)
    dicts = (regression.regression(x, y, k) for x, y, k in xy_iter)
    df = pd.DataFrame(dicts)

    print('Writing to:', out_fname)
    df.to_csv(out_fname, index=False)
    print('Done')


def main():
    setproctitle('kwas')

    arg_parser = argparse.ArgumentParser()
    arg_parser.add_argument('-i', required=True, help="Input HAS file")
    arg_parser.add_argument('-o', required=True, help="Output file")
    arg_parser.add_argument('-c', required=True, help="Covariates file")
    arg_parser.add_argument('-s', required=True, help="Samples file")
    args: argparse.Namespace = arg_parser.parse_args()

    run(args.s, args.i, args.o, args.c)


if __name__ == '__main__':
    main()
