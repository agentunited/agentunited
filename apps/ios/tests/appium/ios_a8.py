from datetime import datetime
from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Channels').click()
      d.find_element('xpath', "//XCUIElementTypeStaticText[contains(@label,'#general') or @label='general']").click()
      msg = f'iOS channel test {datetime.utcnow().strftime("%H%M%S")}'
      assert_visible(d, 'Message input').send_keys(msg)
      assert_visible(d, 'Send').click()
      d.find_element('xpath', f"//XCUIElementTypeStaticText[@label='{msg}']")
      return ok('Channel send works')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
