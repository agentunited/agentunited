from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Message input').send_keys('Hello from iOS QA')
      assert_visible(d, 'Send').click()
      d.find_element('xpath', "//XCUIElementTypeStaticText[@label='Hello from iOS QA']")
      return ok('DM send message appears in thread')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
