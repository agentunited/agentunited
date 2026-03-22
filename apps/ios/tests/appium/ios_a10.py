from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Message input').send_keys('Test unavailable')
      assert_visible(d, 'Send').click()
      d.implicitly_wait(32)
      d.find_element('xpath', "//XCUIElementTypeStaticText[contains(@label,'unavailable')]")
      return ok('Unavailable indicator shown')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
