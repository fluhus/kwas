#define PY_SSIZE_T_CLEAN
#include <Python.h>

#include <algorithm>
#include <cmath>
#include <cstring>
#include <iostream>
#include <vector>

#define TEST_TWO_SIDED "two-sided"
#define TEST_GREATER "greater"
#define TEST_LESS "less"

// Returns an informative ValueError if x is negative.
#define ERR_IF_NEGATIVE(x)                                      \
  if ((x) < 0) {                                                \
    PyErr_Format(PyExc_ValueError, #x " is negative: %i", (x)); \
    return NULL;                                                \
  }

// Stores log-factorials.
std::vector<double> facs = {0};

// Fills the log-factorials cache up to i.
void fill_facs(int i) {
  for (int j = facs.size(); j <= i; j++) {
    facs.push_back(facs[j - 1] + log(j));
  }
}

// Returns the value of a single contingency table configuration.
double fisher_single_step(int a, int b, int c, int d, int n) {
  return exp(facs[a + b] + facs[c + d] + facs[a + c] + facs[b + d] - facs[n] -
             facs[a] - facs[b] - facs[c] - facs[d]);
}

// Returns the p-value of a "greater" test (a getting a higher value).
double fisher_greater(int a, int b, int c, int d, int n) {
  double p = 0;
  while (b >= 0 && c >= 0) {
    p += fisher_single_step(a, b, c, d, n);
    a++;
    b--;
    c--;
    d++;
  }
  return p;
}

// Returns the p-value of a "less" test (a getting a lower value).
double fisher_less(int a, int b, int c, int d, int n) {
  double p = 0;
  while (a >= 0 && d >= 0) {
    p += fisher_single_step(a, b, c, d, n);
    a--;
    b++;
    c++;
    d--;
  }
  return p;
}

// Returns the p-value of a "two-sided" test (how likely to be less likely).
double fisher_two_sided(int a, int b, int c, int d, int n) {
  double orig_p = fisher_single_step(a, b, c, d, n);
  int orig_a = a;

  double p = orig_p;

  // First side, make a low.
  int diff = std::min(a, d);
  a -= diff;
  b += diff;
  c += diff;
  d -= diff;
  while (a < orig_a) {
    double curr_p = fisher_single_step(a, b, c, d, n);
    if (curr_p > orig_p) {  // Event is more probable than input.
      break;
    }
    p += curr_p;
    a++;
    b--;
    c--;
    d++;
  }

  // Second side, make a high.
  diff = std::min(b, c);
  a += diff;
  b -= diff;
  c -= diff;
  d += diff;
  while (a > orig_a) {
    double curr_p = fisher_single_step(a, b, c, d, n);
    if (curr_p > orig_p) {  // Event is more probable than input.
      break;
    }
    p += curr_p;
    a--;
    b++;
    c++;
    d--;
  }

  return p;
}

// API function for fisher testing.
PyObject* fisher(PyObject* self, PyObject* args) {
  int a, b, c, d;
  const char* s = TEST_TWO_SIDED;
  if (!PyArg_ParseTuple(args, "iiii|s", &a, &b, &c, &d, &s)) {
    return NULL;
  }
  ERR_IF_NEGATIVE(a);
  ERR_IF_NEGATIVE(b);
  ERR_IF_NEGATIVE(c);
  ERR_IF_NEGATIVE(d);

  double odr;
  if (b * c == 0) {
    if (a * d == 0) {
      odr = NAN;  // 0/0
    } else {
      odr = INFINITY;  // X/0
    }
  } else {
    odr = ((double)a) * ((double)d) / ((double)b) / ((double)c);
  }

  int n = a + b + c + d;
  fill_facs(n);

  double p;
  if (strcmp(s, TEST_GREATER) == 0) {
    p = fisher_greater(a, b, c, d, n);
  } else if (strcmp(s, TEST_LESS) == 0) {
    p = fisher_less(a, b, c, d, n);
  } else if (strcmp(s, TEST_TWO_SIDED) == 0) {
    p = fisher_two_sided(a, b, c, d, n);
  } else {
    // Creating a python string for nice formatting.
    PyObject* str = PyUnicode_FromString(s);
    PyErr_Format(PyExc_ValueError, "invalid test type: %R", str);
    Py_DECREF(str);
    return NULL;
  }

  return Py_BuildValue("dd", odr, p);
}

// API function for clearing the log-factorial cache.
PyObject* reset(PyObject* self, PyObject* args) {
  if (!PyArg_ParseTuple(args, "")) {
    return NULL;
  }
  std::vector<double>().swap(facs);
  facs.push_back(0);
  Py_RETURN_NONE;
}

static PyMethodDef methods[] = {
    {"fisher", fisher, METH_VARARGS, "A fisher test."},
    {"reset", reset, METH_VARARGS, "Clears the factorial cache."},
    {NULL, NULL, 0, NULL}};

static struct PyModuleDef module = {
    PyModuleDef_HEAD_INIT,
    "myfisher",
    "Fisher's exact test implementation.",
    -1,
    methods,
};

PyMODINIT_FUNC PyInit_myfisher(void) {
  return PyModule_Create(&module);
}
