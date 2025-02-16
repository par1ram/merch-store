import http from 'k6/http'
import { check } from 'k6'

// Параметры теста
export const options = {
	scenarios: {
		main_scenario: {
			// Генерируем постоянную скорость (RPS) = 1000 запросов/с.
			executor: 'constant-arrival-rate',
			rate: 1000, // 1000 итераций (запросов) в 1 секунду => 1k RPS
			timeUnit: '1s', // измеряем rate в запросах/сек
			duration: '1m', // общая длительность 1 минута (пример)
			preAllocatedVUs: 200, // заранее выделенное число VU (подбирайте эмпирически)
			maxVUs: 1000, // максимальное число VU (на случай пиков)
		},
	},

	// Пороговые значения (SLI) для времени ответа и успешности
	thresholds: {
		// 99.99% запросов должны быть быстрее 50 мс:
		'http_req_duration{scenario:main_scenario}': ['p(99.99)<50'],
		// Доля неуспешных запросов (ошибок) < 0.0001 => 0.01% (значит 99.99% успешных).
		'http_req_failed{scenario:main_scenario}': ['rate<0.0001'],
	},
}

export default function () {
	// Пример эндпоинта; замените на ваш:
	const id = Math.floor(Math.random() * 100000)
	const url = `http://localhost:8080/api/info`
	const res = http.get(url, {
		headers: {
			Authorization:
				'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk3NDU3OTAsImlhdCI6MTczOTY1OTM5MCwidXNlcl9pZCI6MSwidXNlcm5hbWUiOiJ2bGFkaWsifQ.XQnC9ugsEZVcwRCRJo03bX78vhLIkHmpGrxJrNBZ02k',
		},
	})

	// Проверяем, что статус == 200
	check(res, {
		'status is 200': (r) => r.status === 200,
	})
}
