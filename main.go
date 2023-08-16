package main

import (
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/op/go-logging"
	"github.com/sg3des/argum"
	"github.com/tarm/serial"
)

var log = logging.MustGetLogger("ELTIS")

var MsgInit = []byte{0x7F, 0x7F, 0x0A, 0x01}
var MsgOpen = []byte{0x7F, 0x40, 0x06, 0x0f}

var args struct {
	Device string `help:"path to device" default:"/dev/ttyACM0"`
	Listen string `help:"http listen address" default:":6976"`
}

func init() {
	argum.MustParse(&args)

	logging.SetBackend(logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stderr, "", 0),
		logging.MustStringFormatter(`%{color}[%{shortfile}] %{message}%{color:reset}`),
	))
}

func main() {
	ctrl, err := NewController(args.Device)
	if err != nil {
		log.Fatal(err)
	}

	// initialize the driver mainly to check that the device is available
	if port, err := ctrl.dial(); err == nil {
		if err := ctrl.write(port, MsgInit); err != nil {
			log.Errorf("failed to initialize driver: %v", err)
		}
	} else {
		// INFO: maybe a fatal error is needed here?
		log.Error("connection failed: %v", err)
	}

	// initialize web server
	app := fiber.New(fiber.Config{ErrorHandler: ErrHandler})
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", ctrl.Open)
	app.Post("/", ctrl.Open)

	if err := app.Listen(args.Listen); err != nil {
		log.Fatal(err)
	}
}

func ErrHandler(c *fiber.Ctx, err error) error {
	code := 500

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	} else {
		err = fiber.NewError(code, err.Error())
	}

	return c.Status(code).JSON(err)
}

type Controller struct {
	conf *serial.Config

	sync.Mutex
}

func NewController(device string) (ctrl *Controller, err error) {
	ctrl = &Controller{
		conf: &serial.Config{
			Name:        device,
			Baud:        9600,
			ReadTimeout: 5 * time.Second,
		},
	}

	return ctrl, nil
}

func (ctrl *Controller) Open(c *fiber.Ctx) error {
	ctrl.Lock()
	defer ctrl.Unlock()

	log.Debug("open...")

	// open port
	port, err := ctrl.dial()
	if err != nil {
		log.Error("connection failed:", err)
		return err
	}

	// init driver
	if err := ctrl.write(port, MsgInit); err != nil {
		log.Error("driver initialize failed", err)
		return err
	}

	// read initialize driver response
	if _, err := ctrl.read(port); err != nil {
		log.Error("read failed:", err)
	}

	// open door
	if err := ctrl.write(port, MsgOpen); err != nil {
		log.Error("open failed:", err)
		return err
	}

	port.Close()

	log.Debug("done")

	return c.JSON(fiber.Map{"success": "true"})
}

func (ctrl *Controller) dial() (port *serial.Port, err error) {
	port, err = serial.OpenPort(ctrl.conf)
	if err != nil {
		return
	}

	return
}

func (ctrl *Controller) write(port *serial.Port, cmd []byte) (err error) {
	msg := make([]byte, 30)
	copy(msg, cmd)

	_, err = port.Write(msg)
	return
}

func (ctrl *Controller) read(port *serial.Port) (resp []byte, err error) {
	buf := make([]byte, 30)

	n, err := port.Read(buf)
	return buf[:n], err
}
