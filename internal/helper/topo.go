package helper

type Task struct {
	ID           string
	Dependencies []string
}

type Graph struct {
	Tasks         map[string]*Task
	AdjacencyList map[string][]string
	InDegree      map[string]int
	Visited       map[string]bool
}

func NewGraph(tasks []Task) *Graph {
	graph := &Graph{
		Tasks:         make(map[string]*Task),
		AdjacencyList: make(map[string][]string),
		InDegree:      make(map[string]int),
		Visited:       make(map[string]bool),
	}

	for _, task := range tasks {
		graph.AddTask(task)
	}

	return graph
}

func (g *Graph) AddTask(task Task) {
	g.Tasks[task.ID] = &task
	g.Visited[task.ID] = false
	for _, dep := range task.Dependencies {
		g.AdjacencyList[dep] = append(g.AdjacencyList[dep], task.ID)
		g.InDegree[task.ID]++
	}
	if _, exists := g.InDegree[task.ID]; !exists {
		g.InDegree[task.ID] = 0
	}
}

func (g *Graph) TopologicalSort() []string {
	var queue []string
	for id, degree := range g.InDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)

		for _, adj := range g.AdjacencyList[id] {
			g.InDegree[adj]--
			if g.InDegree[adj] == 0 {
				queue = append(queue, adj)
			}
		}
	}

	return order
}

func GroupTasks(tasks []Task) [][]string {
	graph := NewGraph(tasks)
	sortedTasks := graph.TopologicalSort()

	groups := [][]string{}
	groupMap := make(map[string]int)

	for _, taskID := range sortedTasks {
		maxGroup := -1
		for _, dep := range graph.Tasks[taskID].Dependencies {
			if groupMap[dep] > maxGroup {
				maxGroup = groupMap[dep]
			}
		}
		groupIndex := maxGroup + 1
		if groupIndex == len(groups) {
			groups = append(groups, []string{})
		}
		groups[groupIndex] = append(groups[groupIndex], taskID)
		groupMap[taskID] = groupIndex
	}

	return groups
}
