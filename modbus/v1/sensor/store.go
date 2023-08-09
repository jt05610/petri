package sensor

import (
	"context"
	"core/sensor/v1/sensor"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type store struct {
	db *gorm.DB
	UnimplementedStoreServer
}

func (s *store) GetSensor(_ context.Context, request *GetSensorRequest) (*GetSensorResponse, error) {
	var result ModbusSensor
	err := s.db.First(&result, request.Id).Error
	if err != nil {
		return nil, err
	}
	return &GetSensorResponse{
		Device: &result,
	}, nil
}

func (s *store) ListSensor(_ context.Context, _ *sensor.Empty) (*ListSensorResponse, error) {
	var result []*ModbusSensor
	err := s.db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return &ListSensorResponse{
		Devices: result,
	}, nil
}

func (s *store) AddSensor(_ context.Context, request *AddSensorRequest) (*AddSensorResponse, error) {
	ax := &ModbusSensor{
		Sensor: &sensor.Sensor{
			Id:          uuid.New().String(),
			Name:        &request.Name,
			Calibration: request.Calibration,
		},
		RegisterMap: request.RegisterMap,
	}
	err := s.db.Create(ax).Error
	if err != nil {
		return nil, err
	}
	return &AddSensorResponse{
		Device: ax,
	}, nil
}

func (s *store) RemoveSensor(_ context.Context, request *RemoveSensorRequest) (*RemoveSensorResponse, error) {
	err := s.db.Delete(&ModbusSensor{}, request.Id).Error
	if err != nil {
		return nil, err
	}
	return &RemoveSensorResponse{
		Device: nil,
	}, nil
}

func (s *store) UpdateSensor(_ context.Context, request *UpdateSensorRequest) (*UpdateSensorResponse, error) {
	var ax ModbusSensor
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	if request.Name != "" {
		ax.Sensor.Name = &request.Name
	}
	if request.Calibration != nil {
		ax.Sensor.Calibration = request.Calibration
	}
	if request.RegisterMap != nil {
		ax.RegisterMap = request.RegisterMap
	}
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateSensorResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateName(_ context.Context, request *UpdateNameRequest) (*UpdateNameResponse, error) {
	var ax ModbusSensor
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	ax.Sensor.Name = &request.Name
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateNameResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateCalibration(_ context.Context, request *UpdateCalibrationRequest) (*UpdateCalibrationResponse, error) {
	var ax ModbusSensor
	err := s.db.First(&ax, request.Id).Error
	if err != nil {
		return nil, err
	}
	ax.Sensor.Calibration = request.Calibration
	err = s.db.Save(&ax).Error
	if err != nil {
		return nil, err
	}
	return &UpdateCalibrationResponse{
		Device: &ax,
	}, nil
}

func (s *store) UpdateRegMap(_ context.Context, request *UpdateRegMapRequest) (*UpdateRegMapResponse, error) {
	var ax ModbusSensor
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
