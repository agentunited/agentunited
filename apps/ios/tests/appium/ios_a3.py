from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver()
    try:
      assert_visible(d, 'Display name')
      d.find_element('accessibility id', 'Display name').send_keys('iOS QA Tester')
      d.find_element('accessibility id', 'Password').send_keys('QaTestiOS123!')
      d.find_element('accessibility id', 'Confirm password').send_keys('QaTestiOS123!')
      d.find_element('accessibility id', 'Join workspace').click()
      assert_visible(d, 'Messages', timeout=12)
      return ok('Signup lands on Messages tab')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
