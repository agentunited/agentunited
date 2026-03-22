from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Channels').click()
      d.find_element('xpath', "//XCUIElementTypeStaticText[contains(@label,'#') or @label='general']")
      return ok('Channel list visible')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
