from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver(no_reset=True)
    try:
      assert_visible(d, 'Messages').click()
      # basic smoke: list exists post-refresh gesture attempt
      el = assert_visible(d, 'Messages')
      d.execute_script('mobile: dragFromToForDuration', {
          'duration': 0.4, 'fromX': 200, 'fromY': 220, 'toX': 200, 'toY': 620
      })
      assert el.is_displayed()
      return ok('Pull-to-refresh gesture executed without crash')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
