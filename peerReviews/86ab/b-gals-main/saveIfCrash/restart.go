package saveIfCrash

import (
	"encoding/json"
	"os"
	"heis/order"

)

const cabOrderFilePermissions = 0644

func cabOrdersPath(id string) string {
    return "cab_orders_" + id + ".json"
}

//saves caborders to disk

func SaveCabOrders(id string, orders []order.Order) error {
    data, err := json.Marshal(orders)
    if err != nil {
        return err
    }
    return os.WriteFile(cabOrdersPath(id), data, cabOrderFilePermissions)
}

//caborders read back from disk after restart

func LoadCabOrders(id string, numFloors int) []order.Order {
    data, err := os.ReadFile(cabOrdersPath(id))
    if err != nil {
        return order.NewCabOrders(numFloors)
    }
    var orders []order.Order
    if err := json.Unmarshal(data, &orders); err != nil {
        return order.NewCabOrders(numFloors)
    }
    return orders
}