/**
 *  Neurago is a Go Library for creating, training, and saving
 *  neural networks.
 */
package neurago
import "crypto/rand"
// Each perceptron will have 1 or more Connectors, connectors
// specify an in channel to read input from (Activate) and
// an out channel to send the results to (Result).
// 
// The out perceptrons will be assigned a channel to get information
// back out.
// 
// The Reward channel recieves an amount as reward (positive) or punishment
// (negative).
const (
    KILLED = iota
    KILLING
    REMOVED
)
type Connector32 struct {
    in chan []float32
    out chan float32
    reward chan float32
    kill chan bool
    err chan int
    id int
}
type ActivationCallback func(sum float32, out chan<- float32, err chan<- int) float32
type RewardCallback func(rw float32,err chan<-)
// We don't really care how you specifically implement
// your perceptron, and we don't want to limit
// ourselves to a single kind of implementation
type Perceptron32 interface {
    AddConnector(id, c Connector) (bool,error)
    RemoveConnector(id int) (bool,error)
    HasConnection(id int) bool
    Activate(sum float32, out chan<- float32, err chan<- int)
    Reward(rw float32, err chan<-)
    SetActivationCallback(cb ActivationCallback)
    SetRewardCallback(cb RewardCallback)
    // because Frankenstein is funny in this context
    // https://youtu.be/0VkrUG3OrPc
    ItsALive() 
}
// This is the basic implementation of Neuron for
// a Neural Network of any composition. You can
// implement your own, so long as it will
// satisfy the Perceptron interface
type Neuron struct {
    connections map[int]Connector32
    bias float32
    weights []float32
    dead bool
    activation_callback ActivationCallback
    reward_callback RewardCallback
}
// Because fucking programmers. Say you want a
// random number and get an earful about pseudo random
// and really kinda sorta random. For fuck's sake, I
// just want a random number!
func Random32() (int,error) {
    b := []byte{0}
    if _, err := rand.Reader.Read(b); err != nil {
        return nil,err
    }
    return rand.New( rand.NewSource( b[0] + time.Now().UnixNano() ) )
}
func (Neuron* p) HasConnection(id int) bool {
    if p.connections[id] {
        return true
    }
    return false
}
func (Neuron* p) AddConnector(id int,c Connector32) (bool,error) {
    if p.HasConnection(id) {
        return false,nil
    }
    p.connections[id] = c
    return true,nil
}
func (Neuron* p) RemoveConnector(id int) (bool,error) {
    if p.HasConnection(id) {
        return false,nil
    }
    p.connections[id] = nil
    return true,nil
}
func (Neuron* p) SetActivationCallback(cb ActivationCallback) {
    p.activation_callback = cb;
}
func (Neuron* p) Activate(sum float32,out chan<- float32, err chan<- int) {
    p.activation_callback(sum,out,err)
}
func (Neuron* p) SetRewardCallback(cb RewardCallback) {
    p.reward_callback = cb;
}
func (Neuron* p) Reward(rw float32, err chan<- int) {
    p.reward_callback(rw,err)
}
func (Neuron* p) ItsAlive() {
    for id,con := range p.connections {
        go func() {
            // we recive inputs on this connection
            for p.dead != true && p.HasConnection(id) {
                inputs []float32
                sum float32
                // we read inputs
                select {
                        case inputs = <- con.in:
                            // this can be avoided by restricting the
                            // inputs and initializing the weights to
                            // your own satisfaction
                            for len(inputs) > len(p.weights) {
                                rv,err := Random32()
                                if err != nil {
                                    con.err <- err
                                    return
                                }
                                p.weights = append( p.weights, rv )
                            }
                            // we make the output
                            for i,v := range inputs {
                                sum += v * p.weights[i]
                            }
                            sum = sum * p.bias
                            // we distribute the results to all
                            // connections
                            for _,ocons := range p.connections {
                                p.Activate(sum,ocons.out,ocons.err)
                            }
                        case shouldDie := <- con.kill:
                            p.dead = true
                            con.err <- KILLED
                            return
                        case rw := <- con.reward:
                            p.Reward(rw,con.err)
                }
            }
        } 
    }
}
func CreateNeuron() (Neuron,error) {
    p Neuron;
    r,err := Random32()
    if err != nil {
        return nil,err
    }
    p.bias = r.Float32;
    return p
}
func CreateConnector32(id int, in chan float32, out chan float32, reward chan float32,kill chan bool, err chan int) (Connector32,error) {
    if in == nil {
        in = make(chan float32)
    }
    if out == nil {
        out = make(chan float32)
    }
    if reward == nil {
        reward = make(chan float32)
    }
    if kill == nil {
        kill = make(chan bool)
    }
    if err == nil {
        err = make(chan int)
    }
    c Connector32;
    c.in = in
    c.out = out
    c.reward = reward
    c.kill = kill
    c.err = err
    return c,nil
}
