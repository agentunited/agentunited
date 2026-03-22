from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Messages').click()
      assert_visible(d, 'New message').click()
      # loose match by common agent label fallback
      _ = d.find_element('xpath', "//XCUIElementTypeStaticText[contains(@label,'empire') or contains(@label,'Empire')]")
      return ok('Agent visible in New Message modal')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
