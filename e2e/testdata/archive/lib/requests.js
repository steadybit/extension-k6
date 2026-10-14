import http from 'k6/http';

export function visit(url) {
	return http.get(url);
}
