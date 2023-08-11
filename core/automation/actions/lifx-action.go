package actions

type LifxAction struct {
	label string
}

func NewLifxAction(label string) *LifxAction {
	return &LifxAction{label: label}
}
func (l *LifxAction) IsConnected() (bool, error) {

	// bulbs, _ := golifx.LookupBulbs()

	// if len(bulbs) == 0 {
	// 	return false, errors.New("no lights found")
	// }
	// for _, bulb := range bulbs {

	// 	bl, err := bulb.GetLabel()
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	if bl == l.label {
	// 		return true, nil
	// 	}
	// }
	return false, nil
}
