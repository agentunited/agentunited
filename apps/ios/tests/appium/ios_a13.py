from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Profile').click()
      assert_visible(d, 'Edit display name').click()
      field = assert_visible(d, 'Display name')
      field.clear()
      field.send_keys('iOS QA Updated')
      assert_visible(d, 'Save').click()
      assert_visible(d, 'iOS QA Updated', timeout=8)
      return ok('Display name updated')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
