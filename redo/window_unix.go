// 7 july 2014

package ui

import (
	"unsafe"
)

// #include "gtk_unix.h"
// extern gboolean windowClosing(GtkWidget *, GdkEvent *, gpointer);
import "C"

type window struct {
	widget	*C.GtkWidget
	container	*C.GtkContainer
	bin		*C.GtkBin
	window	*C.GtkWindow

	closing	*event
}

func newWindow(title string, width int, height int) *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			widget := C.gtk_window_new(C.GTK_WINDOW_TOPLEVEL)
			ctitle := togstr(title)
			defer freegstr(ctitle)
			w := &window{
				widget:		widget,
				container:		(*C.GtkContainer)(unsafe.Pointer(widget)),
				bin:			(*C.GtkBin)(unsafe.Pointer(widget)),
				window:		(*C.GtkWindow)(unsafe.Pointer(widget)),
				closing:		newEvent(),
			}
			C.gtk_window_set_title(w.window, ctitle)
			g_signal_connect(
				C.gpointer(unsafe.Pointer(w.window)),
				"delete-event",
				C.GCallback(C.windowClosing),
				C.gpointer(unsafe.Pointer(w)))
			// TODO size
			// TODO content
			c <- w
		},
		resp:		c,
	}
}

func (w *window) SetControl(control Control) *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			// TODO unparent
			// TODO reparent
			c <- struct{}{}
		},
		resp:		c,
	}
}

func (w *window) Title() *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			c <- fromgstr(C.gtk_window_get_title(w.window))
		},
		resp:		c,
	}
}

func (w *window) SetTitle(title string) *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			ctitle := togstr(title)
			defer freegstr(ctitle)
			C.gtk_window_set_title(w.window, ctitle)
			c <- struct{}{}
		},
		resp:		c,
	}
}


func (w *window) Show() *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			C.gtk_widget_show_all(w.widget)
			c <- struct{}{}
		},
		resp:		c,
	}
}

func (w *window) Hide() *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			C.gtk_widget_hide(w.widget)
			c <- struct{}{}
		},
		resp:		c,
	}
}

func (w *window) Close() *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			C.gtk_widget_destroy(w.widget)
			c <- struct{}{}
		},
		resp:		c,
	}
}

func (w *window) OnClosing(e func(c Doer) bool) *Request {
	c := make(chan interface{})
	return &Request{
		op:		func() {
			w.closing.setbool(e)
			c <- struct{}{}
		},
		resp:		c,
	}
}

//export windowClosing
func windowClosing(wid *C.GtkWidget, e *C.GdkEvent, data C.gpointer) C.gboolean {
	w := (*window)(unsafe.Pointer(data))
	close := w.closing.fire()
	if close {
		return C.GDK_EVENT_PROPAGATE		// will do gtk_widget_destroy(), which is what we want (thanks ebassi in irc.gimp.net/#gtk+)
	}
	return C.GDK_EVENT_STOP				// keeps window alive
}
