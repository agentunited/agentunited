from common import create_driver, assert_visible, ok, fail

def run_test():
    d = create_driver()
    try:
      d.execute_script('mobile: openUrl', {'url': 'agentunited://invite?token=inv_test_xxx&instance=https://w7386887466.tunnel.agentunited.ai'})
      assert_visible(d, 'Join workspace', timeout=8)
      return ok('Invite screen opened via deep link')
    except Exception as e:
      return fail(str(e))
    finally:
      d.quit()
