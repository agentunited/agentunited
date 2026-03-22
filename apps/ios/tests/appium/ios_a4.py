from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Profile').click()
      assert_visible(d, 'iOS QA Tester', timeout=6)
      return ok('Display name visible in profile')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
