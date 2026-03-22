from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Profile').click()
      assert_visible(d, 'Sign out').click()
      assert_visible(d, 'Accept an invite', timeout=8)
      return ok('Sign out returns to welcome')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
