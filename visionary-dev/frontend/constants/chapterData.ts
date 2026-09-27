export interface SubsectionContent {
  id: string
  title: string
  bullets?: string[]
  heading?: string
  items?: {
    title: string
    points: string[]
  }[]
  note?: string
}

export interface Section {
  id: string
  title: string
  subsections?: string[]
  subsectionContent?: SubsectionContent[]
  content: {
    bullets?: string[]
    heading?: string
    items?: {
      title: string
      points: string[]
    }[]
    note?: string
  }
}

export const sections: Section[] = [
  {
    id: '6.1',
    title: '6.1 Introduction',
    subsections: [
      '6.1.1 What kind of motion can a rigid body have?',
      '6.1.2 A large class of problems with extended bodies',
    ],
    subsectionContent: [
      {
        id: '6.1.1',
        title: '6.1.1 What kind of motion can a rigid body have?',
        bullets: [
          'A rigid body can undergo three types of motion: pure translation, pure rotation, and a combination of both.',
          'In pure translation, every particle of the body moves with the same velocity.',
          'In pure rotation, particles move in circles about a fixed axis.',
        ],
      },
      {
        id: '6.1.2',
        title: '6.1.2 A large class of problems with extended bodies',
        bullets: [
          'Many real-world problems involve extended bodies rather than point masses.',
          'Extended bodies have size and shape, unlike point particles.',
          'The motion of extended bodies can be analyzed using the concept of center of mass and rotational dynamics.',
        ],
      },
    ],
    content: {
      bullets: [
        'An extended body is essentially a system of particles, and its overall motion can be understood using the concept of the centre of mass.',
        'Many practical problems are simplified by treating real bodies as rigid bodies, in which the distances between all particles remain constant (even though real bodies actually deform slightly under forces).',
      ],
      heading: 'Types of motion of a rigid body',
      items: [
        {
          title: '1. Pure translational motion',
          points: [
            'All particles move with the same velocity at any instant.',
            'Example: a rectangular block sliding down an inclined plane without rotating.',
          ],
        },
        {
          title: '2. Rotational motion about a fixed axis',
          points: [
            'The rigid body rotates such that every particle moves in a circle lying in a plane perpendicular to the axis, with the centre on the axis.',
            "Examples: ceiling fan, potter's wheel, merry-go-round.",
          ],
        },
      ],
      note: '3. Rolling motion (translation + rotation)',
    },
  },
  {
    id: '6.2',
    title: '6.2 Center of Mass',
    subsections: [
      '6.2.1 Definition and concept',
      '6.2.2 Calculating center of mass',
      '6.2.3 Properties of center of mass',
    ],
    subsectionContent: [
      {
        id: '6.2.1',
        title: '6.2.1 Definition and concept',
        bullets: [
          'The center of mass is the point where the entire mass of a system can be considered to be concentrated.',
          'It is the weighted average position of all the particles in the system.',
          'For a uniform body, the center of mass coincides with the geometric center.',
        ],
      },
      {
        id: '6.2.2',
        title: '6.2.2 Calculating center of mass',
        bullets: [
          'For discrete particles: r_cm = (Σ m_i r_i) / M',
          'For continuous bodies: r_cm = (∫ r dm) / M',
          'The calculation requires knowing the mass distribution.',
        ],
      },
      {
        id: '6.2.3',
        title: '6.2.3 Properties of center of mass',
        bullets: [
          'The center of mass may lie outside the physical body (e.g., a ring).',
          'External forces act as if applied at the center of mass.',
          'The motion of center of mass is independent of internal forces.',
        ],
      },
    ],
    content: {
      bullets: [
        'The center of mass of a system of particles is the point that moves as though all of the mass were concentrated there and all external forces were applied there.',
        'For a system of n particles, the position of center of mass is given by the weighted average of positions.',
      ],
      heading: 'Key Properties',
      items: [
        {
          title: 'Position Vector',
          points: [
            'The position vector of center of mass is independent of the choice of origin.',
            'It depends only on the relative positions and masses of particles.',
          ],
        },
        {
          title: 'Motion',
          points: [
            'The center of mass moves with constant velocity if no external force acts.',
            'Internal forces cannot change the motion of center of mass.',
          ],
        },
      ],
    },
  },
  {
    id: '6.3',
    title: '6.3 Motion of Center of Mass',
    subsections: [
      '6.3.1 Velocity of center of mass',
      '6.3.2 Acceleration of center of mass',
      '6.3.3 Newton\'s second law for systems',
    ],
    subsectionContent: [
      {
        id: '6.3.1',
        title: '6.3.1 Velocity of center of mass',
        bullets: [
          'The velocity of center of mass is v_cm = (Σ m_i v_i) / M',
          'It represents the average velocity of all particles weighted by their masses.',
          'The total momentum equals M v_cm.',
        ],
      },
      {
        id: '6.3.2',
        title: '6.3.2 Acceleration of center of mass',
        bullets: [
          'The acceleration of center of mass is a_cm = F_ext / M',
          'Only external forces affect the acceleration of center of mass.',
          'Internal forces cancel out in pairs due to Newton\'s third law.',
        ],
      },
      {
        id: '6.3.3',
        title: '6.3.3 Newton\'s second law for systems',
        bullets: [
          'F_ext = M a_cm',
          'This is the generalization of Newton\'s second law for a system of particles.',
          'The system behaves as if all mass were concentrated at the center of mass.',
        ],
      },
    ],
    content: {
      bullets: [
        'The velocity of center of mass is the weighted average of velocities of all particles.',
        'The acceleration of center of mass equals the net external force divided by total mass.',
      ],
      heading: 'Important Equations',
      items: [
        {
          title: 'Velocity',
          points: [
            'v_cm = (m₁v₁ + m₂v₂ + ... + mₙvₙ) / M',
            'Where M is the total mass of the system.',
          ],
        },
        {
          title: 'Acceleration',
          points: [
            'a_cm = F_ext / M',
            'This is Newton\'s second law for a system of particles.',
          ],
        },
      ],
    },
  },
  {
    id: '6.4',
    title: '6.4 Linear momentum of a system of particles',
    subsections: [
      '6.4.1 Total momentum',
      '6.4.2 Conservation of momentum',
      '6.4.3 Applications',
    ],
    subsectionContent: [
      {
        id: '6.4.1',
        title: '6.4.1 Total momentum',
        bullets: [
          'Total momentum P = Σ p_i = Σ m_i v_i',
          'It can also be written as P = M v_cm',
          'Momentum is a vector quantity.',
        ],
      },
      {
        id: '6.4.2',
        title: '6.4.2 Conservation of momentum',
        bullets: [
          'If F_ext = 0, then dP/dt = 0, so P = constant',
          'Momentum is conserved in isolated systems.',
          'This principle applies even when internal forces are present.',
        ],
      },
      {
        id: '6.4.3',
        title: '6.4.3 Applications',
        bullets: [
          'Collision problems: momentum before = momentum after',
          'Rocket propulsion: momentum of rocket + exhaust is conserved',
          'Recoil of guns: momentum of bullet + gun is conserved',
        ],
      },
    ],
    content: {
      bullets: [
        'The total linear momentum of a system is the vector sum of momenta of all particles.',
        'The rate of change of total momentum equals the net external force.',
      ],
      heading: 'Conservation Principle',
      items: [
        {
          title: 'Momentum Conservation',
          points: [
            'If net external force is zero, total momentum remains constant.',
            'This is true even if internal forces are present.',
          ],
        },
      ],
    },
  },
  {
    id: '6.5',
    title: '6.5 Vector product of two vectors',
    subsections: [
      '6.5.1 Definition of cross product',
      '6.5.2 Properties of cross product',
      '6.5.3 Applications in physics',
    ],
    subsectionContent: [
      {
        id: '6.5.1',
        title: '6.5.1 Definition of cross product',
        bullets: [
          'A × B is a vector perpendicular to both A and B',
          'Magnitude: |A × B| = |A| |B| sin θ',
          'Direction given by right-hand rule',
        ],
      },
      {
        id: '6.5.2',
        title: '6.5.2 Properties of cross product',
        bullets: [
          'Anti-commutative: A × B = -(B × A)',
          'Not associative: A × (B × C) ≠ (A × B) × C',
          'Distributive: A × (B + C) = A × B + A × C',
        ],
      },
      {
        id: '6.5.3',
        title: '6.5.3 Applications in physics',
        bullets: [
          'Torque: τ = r × F',
          'Angular momentum: L = r × p',
          'Magnetic force: F = q(v × B)',
        ],
      },
    ],
    content: {
      bullets: [
        'The vector product (cross product) of two vectors results in a vector perpendicular to both.',
        'The magnitude equals the product of magnitudes times sine of angle between them.',
      ],
      heading: 'Properties',
      items: [
        {
          title: 'Direction',
          points: [
            'Direction is given by right-hand rule.',
            'A × B is perpendicular to the plane containing A and B.',
          ],
        },
        {
          title: 'Magnitude',
          points: [
            '|A × B| = |A| |B| sin θ',
            'Maximum when vectors are perpendicular, zero when parallel.',
          ],
        },
      ],
    },
  },
  {
    id: '6.6',
    title: '6.6 Angular velocity and its relation with linear velocity',
    subsections: [
      '6.6.1 Angular velocity vector',
      '6.6.2 Relation v = ω × r',
      '6.6.3 Angular acceleration',
    ],
    subsectionContent: [
      {
        id: '6.6.1',
        title: '6.6.1 Angular velocity vector',
        bullets: [
          'Angular velocity ω is a vector along the axis of rotation',
          'Magnitude gives rate of rotation in rad/s',
          'Direction given by right-hand rule',
        ],
      },
      {
        id: '6.6.2',
        title: '6.6.2 Relation v = ω × r',
        bullets: [
          'Linear velocity v = ω × r',
          'v is perpendicular to both ω and r',
          'Magnitude: v = ωr sin θ = ωr⊥',
        ],
      },
      {
        id: '6.6.3',
        title: '6.6.3 Angular acceleration',
        bullets: [
          'Angular acceleration α = dω/dt',
          'Linear acceleration a = α × r + ω × v',
          'Tangential acceleration: a_t = αr',
        ],
      },
    ],
    content: {
      bullets: [
        'Angular velocity describes how fast an object rotates about an axis.',
        'Linear velocity of a point on a rotating body depends on its distance from axis.',
      ],
      heading: 'Relationship',
      items: [
        {
          title: 'Linear and Angular Velocity',
          points: [
            'v = ω × r',
            'Where ω is angular velocity and r is position vector from axis.',
          ],
        },
      ],
    },
  },
  {
    id: '6.7',
    title: '6.7 Torque and angular momentum',
    subsections: [
      '6.7.1 Torque definition',
      '6.7.2 Angular momentum',
      '6.7.3 Relation between torque and angular momentum',
    ],
    subsectionContent: [
      {
        id: '6.7.1',
        title: '6.7.1 Torque definition',
        bullets: [
          'Torque τ = r × F',
          'It is the rotational analogue of force',
          'Causes angular acceleration',
        ],
      },
      {
        id: '6.7.2',
        title: '6.7.2 Angular momentum',
        bullets: [
          'Angular momentum L = r × p = r × mv',
          'For rotation about fixed axis: L = Iω',
          'It is the rotational analogue of linear momentum',
        ],
      },
      {
        id: '6.7.3',
        title: '6.7.3 Relation between torque and angular momentum',
        bullets: [
          'τ = dL/dt',
          'Net torque equals rate of change of angular momentum',
          'If τ = 0, then L is conserved',
        ],
      },
    ],
    content: {
      bullets: [
        'Torque is the rotational analogue of force.',
        'Angular momentum is the rotational analogue of linear momentum.',
      ],
      heading: 'Definitions',
      items: [
        {
          title: 'Torque',
          points: [
            'τ = r × F',
            'Torque causes angular acceleration.',
          ],
        },
        {
          title: 'Angular Momentum',
          points: [
            'L = r × p',
            'Rate of change of angular momentum equals net torque.',
          ],
        },
      ],
    },
  },
  {
    id: '6.8',
    title: '6.8 Equilibrium of a rigid body',
    subsections: [
      '6.8.1 Conditions for equilibrium',
      '6.8.2 Types of equilibrium',
      '6.8.3 Solving equilibrium problems',
    ],
    subsectionContent: [
      {
        id: '6.8.1',
        title: '6.8.1 Conditions for equilibrium',
        bullets: [
          'First condition: Σ F = 0 (translational equilibrium)',
          'Second condition: Σ τ = 0 (rotational equilibrium)',
          'Both conditions must be satisfied simultaneously',
        ],
      },
      {
        id: '6.8.2',
        title: '6.8.2 Types of equilibrium',
        bullets: [
          'Stable equilibrium: system returns to original position when disturbed',
          'Unstable equilibrium: system moves away when disturbed',
          'Neutral equilibrium: system remains in new position when disturbed',
        ],
      },
      {
        id: '6.8.3',
        title: '6.8.3 Solving equilibrium problems',
        bullets: [
          'Draw free body diagram',
          'Choose convenient axis for calculating torques',
          'Apply Σ F = 0 and Σ τ = 0',
        ],
      },
    ],
    content: {
      bullets: [
        'A rigid body is in equilibrium when both net force and net torque are zero.',
        'This ensures no translational or rotational acceleration.',
      ],
      heading: 'Conditions',
      items: [
        {
          title: 'Translational Equilibrium',
          points: [
            'Sum of all forces = 0',
            'No linear acceleration.',
          ],
        },
        {
          title: 'Rotational Equilibrium',
          points: [
            'Sum of all torques = 0',
            'No angular acceleration.',
          ],
        },
      ],
    },
  },
  {
    id: '6.9',
    title: '6.9 Moment of inertia',
    subsections: [
      '6.9.1 Definition and physical meaning',
      '6.9.2 Theorems on moment of inertia',
      '6.9.3 Moment of inertia of common shapes',
    ],
    subsectionContent: [
      {
        id: '6.9.1',
        title: '6.9.1 Definition and physical meaning',
        bullets: [
          'Moment of inertia I = Σ m_i r_i²',
          'It is the rotational analogue of mass',
          'Measures resistance to angular acceleration',
        ],
      },
      {
        id: '6.9.2',
        title: '6.9.2 Theorems on moment of inertia',
        bullets: [
          'Parallel axis theorem: I = I_cm + Md²',
          'Perpendicular axis theorem: I_z = I_x + I_y (for planar bodies)',
          'These theorems simplify calculations',
        ],
      },
      {
        id: '6.9.3',
        title: '6.9.3 Moment of inertia of common shapes',
        bullets: [
          'Thin rod about center: I = ML²/12',
          'Thin rod about end: I = ML²/3',
          'Solid sphere: I = 2MR²/5',
          'Hollow sphere: I = 2MR²/3',
          'Solid cylinder: I = MR²/2',
        ],
      },
    ],
    content: {
      bullets: [
        'Moment of inertia is the rotational analogue of mass.',
        'It depends on mass distribution relative to the axis of rotation.',
      ],
      heading: 'Key Concepts',
      items: [
        {
          title: 'Definition',
          points: [
            'I = Σ mᵢrᵢ²',
            'Where rᵢ is perpendicular distance from axis.',
          ],
        },
        {
          title: 'Parallel Axis Theorem',
          points: [
            'I = I_cm + Md²',
            'Relates moment of inertia about any axis to that about center of mass.',
          ],
        },
      ],
    },
  },
  {
    id: '6.10',
    title: '6.10 Kinematics of rotational motion about a fixed axis',
    subsections: [
      '6.10.1 Angular displacement and velocity',
      '6.10.2 Angular acceleration',
      '6.10.3 Equations of rotational kinematics',
    ],
    subsectionContent: [
      {
        id: '6.10.1',
        title: '6.10.1 Angular displacement and velocity',
        bullets: [
          'Angular displacement θ measured in radians',
          'Angular velocity ω = dθ/dt',
          'Average angular velocity = Δθ/Δt',
        ],
      },
      {
        id: '6.10.2',
        title: '6.10.2 Angular acceleration',
        bullets: [
          'Angular acceleration α = dω/dt',
          'Average angular acceleration = Δω/Δt',
          'Can be positive (speeding up) or negative (slowing down)',
        ],
      },
      {
        id: '6.10.3',
        title: '6.10.3 Equations of rotational kinematics',
        bullets: [
          'ω = ω₀ + αt',
          'θ = ω₀t + ½αt²',
          'ω² = ω₀² + 2αθ',
          'These are analogous to linear kinematic equations',
        ],
      },
    ],
    content: {
      bullets: [
        'Rotational kinematics describes angular position, velocity, and acceleration.',
        'Equations are analogous to linear kinematics.',
      ],
      heading: 'Equations',
      items: [
        {
          title: 'Angular Kinematics',
          points: [
            'ω = ω₀ + αt',
            'θ = ω₀t + ½αt²',
            'ω² = ω₀² + 2αθ',
          ],
        },
      ],
    },
  },
  {
    id: '6.11',
    title: '6.11 Dynamics of rotational motion',
    subsections: [
      '6.11.1 Newton\'s second law for rotation',
      '6.11.2 Work and power in rotational motion',
      '6.11.3 Rolling motion',
    ],
    subsectionContent: [
      {
        id: '6.11.1',
        title: '6.11.1 Newton\'s second law for rotation',
        bullets: [
          'τ = Iα',
          'Net torque = moment of inertia × angular acceleration',
          'Analogous to F = ma',
        ],
      },
      {
        id: '6.11.2',
        title: '6.11.2 Work and power in rotational motion',
        bullets: [
          'Work done by torque: W = τθ',
          'Rotational kinetic energy: KE = ½Iω²',
          'Power: P = τω',
        ],
      },
      {
        id: '6.11.3',
        title: '6.11.3 Rolling motion',
        bullets: [
          'Combination of translation and rotation',
          'Condition for rolling without slipping: v_cm = ωR',
          'Total kinetic energy = ½Mv_cm² + ½Iω²',
        ],
      },
    ],
    content: {
      bullets: [
        'Rotational dynamics relates torque to angular acceleration.',
        'Newton\'s second law for rotation: τ = Iα',
      ],
      heading: 'Key Principles',
      items: [
        {
          title: 'Rotational Equation',
          points: [
            'Net torque = moment of inertia × angular acceleration',
            'Analogous to F = ma for linear motion.',
          ],
        },
        {
          title: 'Work and Energy',
          points: [
            'Rotational kinetic energy = ½Iω²',
            'Work done by torque = τθ',
          ],
        },
      ],
    },
  },
]
