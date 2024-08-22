package monitoring

import (
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

func StatisticsCollectorMiddleware() echo.MiddlewareFunc {
	return echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Registerer: Registry,
		BeforeNext: func(c echo.Context) {
			c.Set("start_time", time.Now())
		},
		AfterNext: func(c echo.Context, err error) {
			start_time := c.Get("start_time").(time.Time)
			request_delay := time.Since(start_time)
			Statistics.Delay.Add(request_delay.Seconds())

			if err != nil {
				Statistics.Requests.WithLabelValues(Unsuccessful).Inc()
			} else {
				Statistics.Requests.WithLabelValues(Successful).Inc()
			}
		},
	})
}
