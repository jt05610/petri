package axis

import (
	"context"
	"core/axis/v1/axis"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type store struct {
	db *gorm.DB
	UnimplementedStoreServer
}

func (s *store) GetAxis(_ context.Context, request *GetAxisRequest) (*GetAxisResponse, error) {
	var result ModbusAxis
	err := s.db.First(&result, request.Id).Error
	if err != nil {
		return nil, err
	}
	return &GetAxisResponse{
		Device: &result,
	}, nil
}

func (s *store) ListAxis(_ context.Context, _ *axis.Empty) (*ListAxisResponse, error) {
	var result []*ModbusAxis
	err := s.db.
		Joins("RegisterMap").
		Joins("Axis").
		Find(&result).Error

	if err != nil {
		return nil, err
	}
	for _, ax := range result {
		err := s.db.Where("register_map_id = ?", ax.RegisterMapId).Find(&ax.RegisterMap.HoldingRegisters).Error
		if err != nil {
			return nil, err
		}
		s.db.Where("register_map_id = ?", ax.RegisterMapId).Find(&ax.RegisterMap.Coils)
		s.db.Where("register_map_id = ?", ax.RegisterMapId).Find(&ax.RegisterMap.InputRegisters)
		s.db.Where("register_map_id = ?", ax.RegisterMapId).Find(&ax.RegisterMap.DiscreteInputs)
		s.db.Where("axis_id = ?", ax.AxisId).Find(&ax.Axis.Calibration)
	}
	return &ListAxisResponse{
		Devices: result,
	}, nil
}

func (s *store) AddAxis(_ context.Context, request *AddAxisRequest) (*AddAxisResponse, error) {
	if request.RegisterMap.Id == "" {
		request.RegisterMap.Id = uuid.New().String()
	}
	ax := &ModbusAxis{
		Id: uuid.New().String(),
		Axis: &axis.Axis{
			Id:          uuid.New().String(),
			UnitId:      &request.UnitId,
			Name:        &request.Name,
			Calibration: request.Calibration,
		},
		RegisterMap:   request.RegisterMap,
		RegisterMapId: request.RegisterMap.Id,
	}
	err := s.db.Create(ax).Error
	if err != nil {
		return nil, err
	}
	return &AddAxisResponse{
		Device: ax,
	}, nil
}

func (s *store) RemoveAxis(_ context.Context, request *RemoveAxisRequest) (*RemoveAxisResponse, error) {
	err := s.db.Delete(&ModbusAxis{}, request.Id).Error
	if err != nil {
		return nil, err
	}
	return &RemoveAxisResponse{
		Device: nil,
	}, nil
}

func (s *store) UpdateAxis(_ context.Context, request *UpdateAxisRequest) (*UpdateAxisResponse, error) {
	var ax ModbusAxis
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	if request.Name != "" {
		ax.Axis.Name = &request.Name
	}
	if request.Calibration != nil {
		ax.Axis.Calibration = request.Calibration
	}
	if request.RegisterMap != nil {
		ax.RegisterMap = request.RegisterMap
	}
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateAxisResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateName(_ context.Context, request *UpdateNameRequest) (*UpdateNameResponse, error) {
	var ax ModbusAxis
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	ax.Axis.Name = &request.Name
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateNameResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateCalibration(_ context.Context, request *UpdateCalibrationRequest) (*UpdateCalibrationResponse, error) {
	var ax ModbusAxis
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	ax.Axis.Calibration = request.Calibration
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateCalibrationResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateRegMap(_ context.Context, request *UpdateRegMapRequest) (*UpdateRegMapResponse, error) {
	var ax ModbusAxis
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	ax.RegisterMap = request.RegisterMap
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateRegMapResponse{
		Device: &ax,
	}, nil
}

func NewStore(db *gorm.DB) StoreServer {
	return &store{db: db}
}
