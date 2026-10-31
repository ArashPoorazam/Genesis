package main

import "math"

// loop logic
func (g *Game) Update() error {
	g.grid.rebuild(g.Prts)

	for i := range g.Prts {
		g.applyInfluence(&g.Prts[i])
	}

	for i := range g.Prts {
		g.Prts[i].move()
		borderCollisionAdjustment(&g.Prts[i])
	}

	return nil
}

// Update particles velocity based on nearby particles.
func (g *Game) applyInfluence(p *Particle) {
	vx, vy := p.VX, p.VY
	cell := g.grid.cell(p.X, p.Y)
	for _, neighbor := range g.grid.neighbors[cell][:g.grid.neighborCount] {
		for otherIndex := g.grid.heads[neighbor]; otherIndex != -1; otherIndex = g.grid.next[otherIndex] {
			other := &g.Prts[otherIndex]
			inner := g.Physics.PT[other.TypeID].InnerInfluenceRadius
			outer := g.Physics.PT[other.TypeID].OuterInfluenceRadius

			dx := wrappedDelta(other.X-p.X, float64(screenWidth))
			dy := wrappedDelta(other.Y-p.Y, float64(screenHeight))
			distanceSquared := dx*dx + dy*dy

			if distanceSquared == 0 || distanceSquared >= g.outerSquared[other.TypeID] {
				continue
			}

			distance := math.Sqrt(distanceSquared)
			relation := g.Physics.Relations[p.TypeID][other.TypeID]
			force := forceCalculation(distance, inner, outer, relation)

			vx += (dx / distance) * force
			vy += (dy / distance) * force
		}
	}

	vx *= damping
	vy *= damping

	speedSquared := vx*vx + vy*vy
	if speedSquared > SpeedLimit*SpeedLimit {
		scale := SpeedLimit / math.Sqrt(speedSquared)
		vx *= scale
		vy *= scale
	}
	p.VX, p.VY = vx, vy
}

func forceCalculation(distance, inner, outer, relation float64) float64 {
	if distance >= outer {
		return 0
	}
	if distance < inner {
		return -repulsionStrength * (1 - distance/inner)
	}

	t := (distance - inner) / (outer - inner)
	weight := 1 - math.Abs(2*t-1)

	return relation * weight
}

func wrappedDelta(delta, size float64) float64 {
	return delta - size*math.Round(delta/size)
}

func wrapPosition(value, size float64) float64 {
	value = math.Mod(value, size)
	if value < 0 {
		value += size
	}
	return value
}

func borderCollisionAdjustment(p *Particle) {
	p.X = wrapPosition(p.X, float64(screenWidth))
	p.Y = wrapPosition(p.Y, float64(screenHeight))
}
