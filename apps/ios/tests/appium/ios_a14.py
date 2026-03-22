from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Profile').click()
      assert_visible(d, 'Open billing').click()
      assert_visible(d, 'Billing', timeout=8)
      return ok('Billing page navigable')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
