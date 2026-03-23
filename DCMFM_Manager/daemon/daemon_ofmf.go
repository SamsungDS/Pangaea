package daemon

import (
	// PKG in mod
	"DCMFM/app/ofmf/handler/v1/fam_pool"
	"DCMFM/config"

	// External PKG
	"github.com/sirupsen/logrus"
)

func OFMF_Initialize(CXLAgent config.CXLAgent) {
	logrus.Debugf("◇◆◇◆Start of OFMF_Initialize()")

	fam_pool.Initialize(CXLAgent)

	// For Debug
	//fam_pool.Test_Bind(CXLAgent)
	//fam_pool.Test_Unbind(CXLAgent)

	logrus.Debugf("◇◆◇◆End of OFMF_Initialize()")
}

func OFMF_Run(CXLAgent config.CXLAgent) {
	logrus.Debugf("◇◆◇◆Start of OFMF_Run()")

	logrus.Debugf("◇◆◇◆End of OFMF_Run()")
}
