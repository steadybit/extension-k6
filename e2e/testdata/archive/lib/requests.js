import http from 'k6/http';

export function visit(url) {
	console.log('k6 e2e: archive ok, imported ../lib/requests.js');
	return http.get(url);
}
