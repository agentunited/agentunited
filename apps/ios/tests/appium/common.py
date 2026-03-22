import os
import time
from appium import webdriver
from appium.options.ios import XCUITestOptions
from selenium.common.exceptions import WebDriverException

UDID = os.getenv('APPIUM_UDID', '00864701-E6E9-4526-8D56-B47C2A883D4D')
BUNDLE_ID = os.getenv('IOS_BUNDLE_ID', 'ai.agentunited.app')


def create_driver(no_reset: bool = False):
    options = XCUITestOptions()
    options.platform_name = 'iOS'
    options.device_name = 'iPhone 16 Pro'
    options.udid = UDID
    options.automation_name = 'XCUITest'
    options.bundle_id = BUNDLE_ID
    options.no_reset = no_reset
    return webdriver.Remote(os.getenv('APPIUM_SERVER', 'http://127.0.0.1:4723'), options=options)


def assert_visible(driver, accessibility_id: str, timeout: int = 8):
    end = time.time() + timeout
    while time.time() < end:
        try:
            el = driver.find_element('accessibility id', accessibility_id)
            if el.is_displayed():
                return el
        except Exception:
            pass
        time.sleep(0.25)
    raise AssertionError(f'Element not visible: {accessibility_id}')


def result(status: str, detail: str = ''):
    return {'status': status, 'detail': detail}


def manual_skip(detail: str):
    return result('SKIP', detail)


def fail(detail: str):
    return result('FAIL', detail)


def ok(detail: str = 'pass'):
    return result('PASS', detail)
