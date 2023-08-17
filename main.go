package main

import (
	"errors"
	"os"
	"path/filepath"
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
var version = "v1.0.0"

var MsgInit = []byte{0x7F, 0x7F, 0x0A, 0x01}
var MsgOpen = []byte{0x7F, 0x40, 0x06, 0x0f}
var door sync.Mutex

var args struct {
	Listen string `help:"http listen address" default:":6976"`
}

func init() {
	argum.Version = version
	argum.MustParse(&args)

	logging.SetBackend(logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stderr, "", 0),
		logging.MustStringFormatter(`%{color}[%{shortfile}] %{message}%{color:reset}`),
	))
}

func main() {
	app := fiber.New(fiber.Config{ErrorHandler: ErrHandler})
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", Open)
	app.Post("/", Open)

	app.Get("/open/:id", Open)
	app.Post("/open/:id", Open)

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

func Open(c *fiber.Ctx) error {
	door.Lock()
	defer door.Unlock()

	id, _ := c.ParamsInt("id", 0)
	msgOpen := append([]byte{}, MsgOpen...)
	msgOpen[1] += uint8(id)

	log.Debug("open:", id)

	ctrl, err := NewControllerAuto()
	if err != nil {
		log.Error("controller failed:", err)
		return err
	}

	log.Debug("file:", ctrl.conf.Name)

	// open port
	port, err := ctrl.dial()
	if err != nil {
		log.Error("connection failed:", err)
		return err
	}
	defer port.Close()

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
	if err := ctrl.write(port, msgOpen); err != nil {
		log.Error("open failed:", err)
		return err
	}

	log.Debug("done")

	return c.JSON(fiber.Map{"success": "true"})
}

//
// controller
//

type Controller struct {
	conf *serial.Config
}

func NewControllerAuto() (ctrl *Controller, err error) {
	files, err := filepath.Glob("/dev/ttyACM*")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("device interface not found")
	}

	return NewController(files[0])
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
