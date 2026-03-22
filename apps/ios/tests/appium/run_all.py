import importlib
import json

TESTS = [
    'ios_a1', 'ios_a2', 'ios_a3', 'ios_a4', 'ios_a5',
    'ios_a6', 'ios_a7', 'ios_a8', 'ios_a9', 'ios_a10',
    'ios_a11', 'ios_a12', 'ios_a13', 'ios_a14', 'ios_a15',
]


def main():
    results = []
    for mod_name in TESTS:
        mod = importlib.import_module(mod_name)
        item = mod_name.upper().replace('_', '-')
        try:
            out = mod.run_test()
        except Exception as e:
            out = {'status': 'FAIL', 'detail': str(e)}
        out['item'] = item
        results.append(out)
        print(f"{item}: {out['status']} - {out.get('detail','')}")

    print('\n=== SUMMARY ===')
    print(json.dumps(results, indent=2))

    fails = [r for r in results if r['status'] == 'FAIL']
    raise SystemExit(1 if fails else 0)


if __name__ == '__main__':
    main()
