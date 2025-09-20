package st7789

// ST7789 command registers
const (
	// Basic commands
	NOP     = 0x00
	SWRESET = 0x01
	RDDID   = 0x04
	RDDST   = 0x09
	SLPIN   = 0x10
	SLPOUT  = 0x11
	PTLON   = 0x12
	NORON   = 0x13
	INVOFF  = 0x20
	INVON   = 0x21
	DISPOFF = 0x28
	DISPON  = 0x29
	CASET   = 0x2A
	RASET   = 0x2B
	RAMWR   = 0x2C
	RAMRD   = 0x2E
	PTLAR   = 0x30
	COLMOD  = 0x3A
	MADCTL  = 0x36

	// MADCTL bits
	MADCTL_MY  = 0x80
	MADCTL_MX  = 0x40
	MADCTL_MV  = 0x20
	MADCTL_ML  = 0x10
	MADCTL_RGB = 0x00
	MADCTL_BGR = 0x08
	MADCTL_MH  = 0x04

	// Extended commands specific to LilyGo devices
	PORCTRL  = 0xB2 // Porch control (was FRMCTR2)
	GCTRL    = 0xB7 // Gate control
	VCOMS    = 0xBB // VCOMS setting
	LCMCTRL  = 0xC0 // LCM control
	VDVVRHEN = 0xC2 // VDV and VRH command enable
	VRHS     = 0xC3 // VRH set
	VDVS     = 0xC4 // VDV setting
	FRCTRL2  = 0xC6 // Frame rate control in normal mode
	PWCTRL1  = 0xD0 // Power control 1
	GMCTRP1  = 0xE0 // Positive voltage gamma control
	GMCTRN1  = 0xE1 // Negative voltage gamma control

	// Brightness control commands (LilyGo optimized)
	WRDISBV = 0x51 // Write display brightness
	WRCTRLD = 0x53 // Write CTRL display
	WRCABC  = 0x55 // Write content adaptive brightness control

	// Read commands
	RDID1 = 0xDA
	RDID2 = 0xDB
	RDID3 = 0xDC
	RDID4 = 0xDD
	GSCAN = 0x45

	// Scroll commands
	VSCRDEF  = 0x33
	VSCRSADD = 0x37

	// Legacy command aliases for compatibility
	FRMCTR1 = 0xB1
	RGBCTRL = 0xB1
	FRMCTR2 = 0xB2
	FRMCTR3 = 0xB3
	INVCTR  = 0xB4
	DISSET5 = 0xB6
	PWCTR1  = 0xC0
	PWCTR2  = 0xC1
	PWCTR3  = 0xC2
	PWCTR4  = 0xC3
	PWCTR5  = 0xC4
	VMCTR1  = 0xC5
	PWCTR6  = 0xFC
)

// Color formats supported by ST7789
const (
	ColorRGB444 = 0b011
	ColorRGB565 = 0b101
	ColorRGB666 = 0b111
)

// Frame rates supported by ST7789 (for FRCTRL2 register)
const (
	FRAMERATE_111 = 0x01
	FRAMERATE_105 = 0x02
	FRAMERATE_99  = 0x03
	FRAMERATE_94  = 0x04
	FRAMERATE_90  = 0x05
	FRAMERATE_86  = 0x06
	FRAMERATE_82  = 0x07
	FRAMERATE_78  = 0x08
	FRAMERATE_75  = 0x09
	FRAMERATE_72  = 0x0A
	FRAMERATE_69  = 0x0B
	FRAMERATE_67  = 0x0C
	FRAMERATE_64  = 0x0D
	FRAMERATE_62  = 0x0E
	FRAMERATE_60  = 0x0F // 60Hz is default
	FRAMERATE_58  = 0x10
	FRAMERATE_57  = 0x11
	FRAMERATE_55  = 0x12
	FRAMERATE_53  = 0x13
	FRAMERATE_52  = 0x14
	FRAMERATE_50  = 0x15
	FRAMERATE_49  = 0x16
	FRAMERATE_48  = 0x17
	FRAMERATE_46  = 0x18
	FRAMERATE_45  = 0x19
	FRAMERATE_44  = 0x1A
	FRAMERATE_43  = 0x1B
	FRAMERATE_42  = 0x1C
	FRAMERATE_41  = 0x1D
	FRAMERATE_40  = 0x1E
	FRAMERATE_39  = 0x1F
)

// Maximum VSYNC scanlines
const MAX_VSYNC_SCANLINES = 254
