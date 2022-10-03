import unittest
from myfisher import fisher
from scipy.stats import fisher_exact


class TestFisher(unittest.TestCase):
    def test_fisher(self):
        tests = [
            [1, 1, 1, 1],
            [1, 0, 0, 1],
            [0, 1, 1, 0],
            [1, 9, 11, 3],
            [0, 10, 12, 2],
            [60, 10, 30, 25],
        ]
        for a, b, c, d in tests:
            for typ in ('two-sided', 'greater', 'less'):
                got = fisher(a, b, c, d, typ)
                want = fisher_exact([[a, b], [c, d]], typ)
                self.assertAlmostEqual(got[0], want[0], 10, f'{(a,b,c,d)}')
                self.assertAlmostEqual(got[1], want[1], 10, f'{(a,b,c,d)}')

    def test_fisher_bad(self):
        self.assertRaises(ValueError, lambda: fisher(-1, 1, 1, 1))
        self.assertRaises(ValueError, lambda: fisher(1, -1, 1, 1))
        self.assertRaises(ValueError, lambda: fisher(1, 1, -1, 1))
        self.assertRaises(ValueError, lambda: fisher(1, 1, 1, -1))
        self.assertRaises(ValueError, lambda: fisher(1, 1, 1, 1, 'aaa'))


if __name__ == '__main__':
    unittest.main()
