import sys
import os
import unittest
import time
from unittest.mock import patch, Mock
sys.path.append(os.path.dirname(__file__) + '/../../python')
from RestClient import RestClient, APIException

class MockResponse:
    pass

# mockup time.strftime within RestClient to return a predetermined value
@patch('RestClient.time', Mock(
    strftime=Mock(return_value='2018-10-20T10:11:12', call_args='%Y-%m-%dT%H:%M:%S%z')
))

class RestClientTest(unittest.TestCase):
    def setUp(self):
        api_user = 42
        api_secret = 'sesame'
        api_url = 'https://unit.local'
        api_realm = 'UNIT'
        self.apiv1 = RestClient(api_user, api_secret, api_url, 1, api_realm)
        self.apiv3 = RestClient(api_user, api_secret, api_url, 3, api_realm)

    def test_timestamp(self):
        self.assertEqual(self.apiv1._get_new_timestamp(), '2018-10-20T10:11:12')

    def test_hash_post(self):
        self.apiv1._set_request_data('getStuff', [1, 'unit'], 'POST')
        self.assertEqual(self.apiv1._get_hash(), '728d7c26b9d3fc42da3d518a4097da81')

    def test_hash_get(self):
        self.apiv3._set_request_data('getStuff?param1=something&param2=somethingelse', [], 'GET')
        self.assertEqual(self.apiv3._get_hash(), 'd41d8cd98f00b204e9800998ecf8427e')

    def test_message_post(self):
        self.apiv1._set_request_data('getStuff', [1, 'unit'], 'POST')
        expected = "POST\n" \
            "728d7c26b9d3fc42da3d518a4097da81\n" \
            "application/json\n" \
            "2018-10-20T10:11:12\n" \
            "[1, \"unit\"]\n" \
            "/api/v1/getStuff"

        self.assertEqual(self.apiv1._get_message(), expected.encode('utf-8'))

    def test_message_get(self):
        self.apiv3._set_request_data('getStuff?param1=something&param2=somethingelse', [], 'GET')
        expected = "GET\n" \
            "d41d8cd98f00b204e9800998ecf8427e\n" \
            "application/json\n" \
            "2018-10-20T10:11:12\n" \
            "\n" \
            "/api/v3/getStuff?param1=something&param2=somethingelse"

        self.assertEqual(self.apiv3._get_message(), expected.encode('utf-8'))

    def test_signature_post(self):
        self.apiv1._set_request_data('getStuff', [1, 'unit'], 'POST')
        self.assertEqual(
            self.apiv1._get_signature(),
            '37aa7ce64e02cca021917606b7932eee3bec7135d1b14432d44018a4e4cc85ad'
        )
    def test_signature_get(self):
        self.apiv3._set_request_data('getStuff?param1=something&param2=somethingelse', [], 'GET')
        self.assertEqual(
            self.apiv3._get_signature(),
            '27c82b98af8498297020fb4524e538f57d0a62df31daf72e9f5c9e07fcde30a9'
        )

    def test_headers_post(self):
        self.apiv1._set_request_data('getStuff', [1, 'unit'], 'POST')
        expected = {
            'Content-Type': 'application/json',
            'Content-MD5': '728d7c26b9d3fc42da3d518a4097da81',
            'Date': '2018-10-20T10:11:12',
            'Authorization': 'UNIT 42:37aa7ce64e02cca021917606b7932eee3bec7135d1b14432d44018a4e4cc85ad',
        }

        self.assertEqual(self.apiv1._get_headers(), expected)

    def test_headers_get(self):
        self.apiv3._set_request_data('getStuff?param1=something&param2=somethingelse', [], 'GET')
        expected = {
            'Content-Type': 'application/json',
            'Content-MD5': 'd41d8cd98f00b204e9800998ecf8427e',
            'Date': '2018-10-20T10:11:12',
            'Authorization': 'UNIT 42:27c82b98af8498297020fb4524e538f57d0a62df31daf72e9f5c9e07fcde30a9',
        }

        self.assertEqual(self.apiv3._get_headers(), expected)

    @patch('RestClient.requests')
    def test_call_ok(self, mock_requests):
        mock_response = MockResponse
        mock_response.status_code = 200
        mock_response.text = '{"succeed":true,"result":"fake response message"}'
        mock_requests.post = Mock(return_value=mock_response)

        ret = self.apiv1.call('getStuff', [1, 'unit'])

        self.assertEqual(ret, 'fake response message')

        mock_requests.post.assert_called_once_with(
            'https://unit.local/api/v1/getStuff',
            data='[1, "unit"]',
            headers={
                'Content-Type': 'application/json',
                'Content-MD5': '728d7c26b9d3fc42da3d518a4097da81',
                'Date': '2018-10-20T10:11:12',
                'Authorization': 'UNIT 42:37aa7ce64e02cca021917606b7932eee3bec7135d1b14432d44018a4e4cc85ad'
            }
        )

        response = self.apiv1.get_http_response()

        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.text, '{"succeed":true,"result":"fake response message"}')

    @patch('RestClient.requests')
    def test_call_not_ok(self, mock_requests):
        mock_response = MockResponse
        mock_response.status_code = 200
        mock_response.text = '{"succeed":false,"message":"feil"}'
        mock_requests.post = Mock(return_value=mock_response)

        error = ''
        try:
            ret = self.apiv1.call('getStuff', [1, 'unit'])
        except APIException as e:
            error = str(e)

        self.assertEqual(error, 'feil')

    @patch('RestClient.requests')
    def test_unexpected_response_500(self, mock_requests):
        mock_response = MockResponse
        mock_response.status_code = 500
        mock_response.text = 'Internal Server Error 500'
        mock_requests.post = Mock(return_value=mock_response)

        error = ''
        try:
            self.apiv1.call('getStuff', [1, 'unit'])
        except APIException as e:
            error = str(e)

        self.assertEqual(error, 'API response with status code 500')


    @patch('RestClient.requests')
    def test_unexpected_response_200(self, mock_requests):
        mock_response = MockResponse
        mock_response.status_code = 200
        mock_response.text = ''
        mock_requests.post = Mock(return_value=mock_response)

        error = ''
        try:
            self.apiv1.call('getStuff', [1, 'unit'])
        except APIException as e:
            error = str(e)

        self.assertEqual(error, 'API did not return a valid json string')

    def test_error_scenarios(self):
        passed = False
        try:
            hash_ = self.apiv1._get_hash()
        except AttributeError as e:
            passed = True

        self.assertTrue(passed, "get_hash fails without set_request")

        passed = False
        try:
            hash_ = self.apiv1._get_message()
        except AttributeError as e:
            passed = True

        self.assertTrue(passed, "get_message fails without set_request")

        passed = False
        try:
            hash_ = self.apiv1._get_signature()
        except AttributeError as e:
            passed = True

        self.assertTrue(passed, "get_signature fails without set_request")

        passed = False
        try:
            hash_ = self.apiv1._get_headers()
        except AttributeError as e:
            passed = True

        self.assertTrue(passed, "get_headers fails without set_request")


if __name__ == '__main__':
    unittest.main()

