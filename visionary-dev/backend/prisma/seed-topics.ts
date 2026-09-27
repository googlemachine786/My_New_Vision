import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

type SubtopicDef = { pointNumber: string; title: string };
type TopicDef = { pointNumber: string; title: string; subtopics?: SubtopicDef[] };
type ChapterData = { bookName: string; chapterName: string; topics: TopicDef[] };

const DATA: ChapterData[] = [
  // ─── Science 6 (common) ──────────────────────────────────────────────────────
  {
    bookName: "Science 6", chapterName: "Chapter 1",
    topics: [
      { pointNumber: "1.1", title: "Sources of Food", subtopics: [
        { pointNumber: "1.1.1", title: "Food from Plants" },
        { pointNumber: "1.1.2", title: "Food from Animals" },
      ]},
      { pointNumber: "1.2", title: "Plant Parts as Food", subtopics: [
        { pointNumber: "1.2.1", title: "Roots, Stems and Leaves" },
        { pointNumber: "1.2.2", title: "Flowers, Fruits and Seeds" },
      ]},
      { pointNumber: "1.3", title: "What Do Animals Eat?", subtopics: [
        { pointNumber: "1.3.1", title: "Herbivores and Carnivores" },
        { pointNumber: "1.3.2", title: "Omnivores" },
      ]},
    ],
  },
  {
    bookName: "Science 6", chapterName: "Chapter 2",
    topics: [
      { pointNumber: "2.1", title: "Nutrients in Food", subtopics: [
        { pointNumber: "2.1.1", title: "Carbohydrates and Fats" },
        { pointNumber: "2.1.2", title: "Proteins and Vitamins" },
      ]},
      { pointNumber: "2.2", title: "Balanced Diet", subtopics: [
        { pointNumber: "2.2.1", title: "Importance of a Balanced Diet" },
        { pointNumber: "2.2.2", title: "Deficiency Diseases" },
      ]},
    ],
  },
  {
    bookName: "Science 6", chapterName: "Chapter 3",
    topics: [
      { pointNumber: "3.1", title: "Variety in Fabrics", subtopics: [
        { pointNumber: "3.1.1", title: "Natural and Synthetic Fibres" },
        { pointNumber: "3.1.2", title: "Plant Fibres" },
      ]},
      { pointNumber: "3.2", title: "Yarn to Fabric", subtopics: [
        { pointNumber: "3.2.1", title: "Weaving" },
        { pointNumber: "3.2.2", title: "Knitting" },
      ]},
    ],
  },
  {
    bookName: "Science 6", chapterName: "Chapter 7",
    topics: [
      { pointNumber: "7.1", title: "Herbs, Shrubs and Trees", subtopics: [
        { pointNumber: "7.1.1", title: "Differences among Herbs Shrubs Trees" },
        { pointNumber: "7.1.2", title: "Climbers and Creepers" },
      ]},
      { pointNumber: "7.2", title: "Stem and Root", subtopics: [
        { pointNumber: "7.2.1", title: "Functions of the Stem" },
        { pointNumber: "7.2.2", title: "Types of Roots" },
      ]},
      { pointNumber: "7.3", title: "Leaf and Flower", subtopics: [
        { pointNumber: "7.3.1", title: "Parts of a Leaf" },
        { pointNumber: "7.3.2", title: "Parts of a Flower" },
      ]},
    ],
  },
  {
    bookName: "Science 6", chapterName: "Chapter 10",
    topics: [
      { pointNumber: "10.1", title: "Standard Units of Measurement", subtopics: [
        { pointNumber: "10.1.1", title: "SI Units" },
        { pointNumber: "10.1.2", title: "Measuring Length Correctly" },
      ]},
      { pointNumber: "10.2", title: "Types of Motion", subtopics: [
        { pointNumber: "10.2.1", title: "Rectilinear and Circular Motion" },
        { pointNumber: "10.2.2", title: "Periodic Motion" },
      ]},
    ],
  },
  {
    bookName: "Science 6", chapterName: "Chapter 12",
    topics: [
      { pointNumber: "12.1", title: "Electric Cell and Bulb", subtopics: [
        { pointNumber: "12.1.1", title: "Structure of an Electric Cell" },
        { pointNumber: "12.1.2", title: "How a Bulb Glows" },
      ]},
      { pointNumber: "12.2", title: "Electric Circuit and Switch", subtopics: [
        { pointNumber: "12.2.1", title: "Open and Closed Circuits" },
        { pointNumber: "12.2.2", title: "Role of a Switch" },
      ]},
      { pointNumber: "12.3", title: "Conductors and Insulators", subtopics: [
        { pointNumber: "12.3.1", title: "Examples of Conductors" },
        { pointNumber: "12.3.2", title: "Examples of Insulators" },
      ]},
    ],
  },

  // ─── Science 7 (common) ──────────────────────────────────────────────────────
  {
    bookName: "Science 7", chapterName: "Chapter 1",
    topics: [
      { pointNumber: "1.1", title: "Mode of Nutrition in Plants", subtopics: [
        { pointNumber: "1.1.1", title: "Autotrophic Nutrition" },
        { pointNumber: "1.1.2", title: "Heterotrophic Nutrition" },
      ]},
      { pointNumber: "1.2", title: "Photosynthesis", subtopics: [
        { pointNumber: "1.2.1", title: "Conditions Needed for Photosynthesis" },
        { pointNumber: "1.2.2", title: "Chlorophyll and Sunlight" },
      ]},
    ],
  },
  {
    bookName: "Science 7", chapterName: "Chapter 6",
    topics: [
      { pointNumber: "6.1", title: "Why Do We Respire?", subtopics: [
        { pointNumber: "6.1.1", title: "Energy from Food" },
        { pointNumber: "6.1.2", title: "Aerobic and Anaerobic Respiration" },
      ]},
      { pointNumber: "6.2", title: "Breathing in Organisms", subtopics: [
        { pointNumber: "6.2.1", title: "Breathing in Humans" },
        { pointNumber: "6.2.2", title: "Breathing in Fish and Insects" },
      ]},
    ],
  },
  {
    bookName: "Science 7", chapterName: "Chapter 9",
    topics: [
      { pointNumber: "9.1", title: "Slow or Fast?", subtopics: [
        { pointNumber: "9.1.1", title: "Speed and Its Measurement" },
        { pointNumber: "9.1.2", title: "Units of Speed" },
      ]},
      { pointNumber: "9.2", title: "Measurement of Time", subtopics: [
        { pointNumber: "9.2.1", title: "Simple Pendulum" },
        { pointNumber: "9.2.2", title: "Time Period" },
      ]},
      { pointNumber: "9.3", title: "Distance-Time Graph", subtopics: [
        { pointNumber: "9.3.1", title: "Plotting the Graph" },
        { pointNumber: "9.3.2", title: "Interpreting the Graph" },
      ]},
    ],
  },
  {
    bookName: "Science 7", chapterName: "Chapter 14",
    topics: [
      { pointNumber: "14.1", title: "Effects of Electric Current", subtopics: [
        { pointNumber: "14.1.1", title: "Heating Effect" },
        { pointNumber: "14.1.2", title: "Magnetic Effect" },
      ]},
      { pointNumber: "14.2", title: "Electromagnet", subtopics: [
        { pointNumber: "14.2.1", title: "Making an Electromagnet" },
        { pointNumber: "14.2.2", title: "Applications of Electromagnets" },
      ]},
    ],
  },

  // ─── Science 8 (common) ──────────────────────────────────────────────────────
  {
    bookName: "Science 8", chapterName: "Chapter 11",
    topics: [
      { pointNumber: "11.1", title: "Force: A Push or a Pull", subtopics: [
        { pointNumber: "11.1.1", title: "Contact and Non-contact Forces" },
        { pointNumber: "11.1.2", title: "Effects of Force" },
      ]},
      { pointNumber: "11.2", title: "Pressure", subtopics: [
        { pointNumber: "11.2.1", title: "Pressure Exerted by Liquids and Gases" },
        { pointNumber: "11.2.2", title: "Atmospheric Pressure" },
      ]},
    ],
  },
  {
    bookName: "Science 8", chapterName: "Chapter 12",
    topics: [
      { pointNumber: "12.1", title: "Force of Friction", subtopics: [
        { pointNumber: "12.1.1", title: "Factors Affecting Friction" },
        { pointNumber: "12.1.2", title: "Sliding and Rolling Friction" },
      ]},
      { pointNumber: "12.2", title: "Increasing and Reducing Friction", subtopics: [
        { pointNumber: "12.2.1", title: "When Friction Is Useful" },
        { pointNumber: "12.2.2", title: "Fluid Friction" },
      ]},
    ],
  },
  {
    bookName: "Science 8", chapterName: "Chapter 13",
    topics: [
      { pointNumber: "13.1", title: "Sound Is Produced by Vibration", subtopics: [
        { pointNumber: "13.1.1", title: "Vibrating Objects and Sound" },
        { pointNumber: "13.1.2", title: "Human Vocal Cords" },
      ]},
      { pointNumber: "13.2", title: "Propagation of Sound", subtopics: [
        { pointNumber: "13.2.1", title: "Sound Needs a Medium" },
        { pointNumber: "13.2.2", title: "Sound in Solids and Liquids" },
      ]},
      { pointNumber: "13.3", title: "Amplitude, Time Period and Frequency", subtopics: [
        { pointNumber: "13.3.1", title: "Loudness and Pitch" },
        { pointNumber: "13.3.2", title: "Audible and Inaudible Sounds" },
      ]},
    ],
  },
  {
    bookName: "Science 8", chapterName: "Chapter 16",
    topics: [
      { pointNumber: "16.1", title: "Laws of Reflection", subtopics: [
        { pointNumber: "16.1.1", title: "Regular and Diffused Reflection" },
        { pointNumber: "16.1.2", title: "Reflected Light Can Reflect Again" },
      ]},
      { pointNumber: "16.2", title: "Human Eye", subtopics: [
        { pointNumber: "16.2.1", title: "Structure of the Eye" },
        { pointNumber: "16.2.2", title: "Care of the Eyes" },
      ]},
    ],
  },

  // ─── Mathematics 6 (common) ───────────────────────────────────────────────────
  {
    bookName: "Mathematics 6", chapterName: "Chapter 1",
    topics: [
      { pointNumber: "1.1", title: "Comparing Numbers", subtopics: [
        { pointNumber: "1.1.1", title: "How Many Numbers Can You Make?" },
        { pointNumber: "1.1.2", title: "Shifting Digits" },
      ]},
      { pointNumber: "1.2", title: "Large Numbers in Practice", subtopics: [
        { pointNumber: "1.2.1", title: "Estimation" },
        { pointNumber: "1.2.2", title: "Roman Numerals" },
      ]},
    ],
  },
  {
    bookName: "Mathematics 6", chapterName: "Chapter 7",
    topics: [
      { pointNumber: "7.1", title: "A Fraction", subtopics: [
        { pointNumber: "7.1.1", title: "Parts of a Fraction" },
        { pointNumber: "7.1.2", title: "Fraction on the Number Line" },
      ]},
      { pointNumber: "7.2", title: "Proper and Improper Fractions", subtopics: [
        { pointNumber: "7.2.1", title: "Proper Fractions" },
        { pointNumber: "7.2.2", title: "Mixed Fractions" },
      ]},
      { pointNumber: "7.3", title: "Equivalent Fractions", subtopics: [
        { pointNumber: "7.3.1", title: "Finding Equivalent Fractions" },
        { pointNumber: "7.3.2", title: "Simplest Form of a Fraction" },
      ]},
    ],
  },
  {
    bookName: "Mathematics 6", chapterName: "Chapter 11",
    topics: [
      { pointNumber: "11.1", title: "Matchstick Patterns", subtopics: [
        { pointNumber: "11.1.1", title: "Idea of a Variable" },
        { pointNumber: "11.1.2", title: "More Matchstick Patterns" },
      ]},
      { pointNumber: "11.2", title: "Algebraic Expressions", subtopics: [
        { pointNumber: "11.2.1", title: "Use of Variables in Common Rules" },
        { pointNumber: "11.2.2", title: "Expressions with Variables" },
      ]},
    ],
  },

  // ─── Mathematics 7 (common) ───────────────────────────────────────────────────
  {
    bookName: "Mathematics 7", chapterName: "Chapter 1",
    topics: [
      { pointNumber: "1.1", title: "Recall of Integers", subtopics: [
        { pointNumber: "1.1.1", title: "Properties of Addition and Subtraction" },
        { pointNumber: "1.1.2", title: "Multiplication of Integers" },
      ]},
      { pointNumber: "1.2", title: "Division of Integers", subtopics: [
        { pointNumber: "1.2.1", title: "Properties of Division" },
        { pointNumber: "1.2.2", title: "Word Problems" },
      ]},
    ],
  },
  {
    bookName: "Mathematics 7", chapterName: "Chapter 6",
    topics: [
      { pointNumber: "6.1", title: "Median and Altitude of a Triangle", subtopics: [
        { pointNumber: "6.1.1", title: "Medians of a Triangle" },
        { pointNumber: "6.1.2", title: "Altitudes of a Triangle" },
      ]},
      { pointNumber: "6.2", title: "Exterior Angle Property", subtopics: [
        { pointNumber: "6.2.1", title: "Exterior Angle of a Triangle" },
        { pointNumber: "6.2.2", title: "Angle Sum Property" },
      ]},
      { pointNumber: "6.3", title: "Pythagoras Property", subtopics: [
        { pointNumber: "6.3.1", title: "Right-Angled Triangles" },
        { pointNumber: "6.3.2", title: "Applying Pythagoras Theorem" },
      ]},
    ],
  },

  // ─── Mathematics 8 (common) ───────────────────────────────────────────────────
  {
    bookName: "Mathematics 8", chapterName: "Chapter 1",
    topics: [
      { pointNumber: "1.1", title: "Properties of Rational Numbers", subtopics: [
        { pointNumber: "1.1.1", title: "Closure Property" },
        { pointNumber: "1.1.2", title: "Commutativity and Associativity" },
      ]},
      { pointNumber: "1.2", title: "Rational Numbers on the Number Line", subtopics: [
        { pointNumber: "1.2.1", title: "Representation of Rational Numbers" },
        { pointNumber: "1.2.2", title: "Finding Rational Numbers Between Two Rationals" },
      ]},
    ],
  },
  {
    bookName: "Mathematics 8", chapterName: "Chapter 6",
    topics: [
      { pointNumber: "6.1", title: "Properties of Square Numbers", subtopics: [
        { pointNumber: "6.1.1", title: "Patterns in Square Numbers" },
        { pointNumber: "6.1.2", title: "Pythagorean Triplets" },
      ]},
      { pointNumber: "6.2", title: "Finding Square Root", subtopics: [
        { pointNumber: "6.2.1", title: "Square Root by Repeated Subtraction" },
        { pointNumber: "6.2.2", title: "Square Root by Long Division" },
      ]},
    ],
  },

  // ─── Science Class 9 – Physics chapters ──────────────────────────────────────
  {
    bookName: "Science Class 9", chapterName: "Motion",
    topics: [
      { pointNumber: "8.1", title: "Describing Motion", subtopics: [
        { pointNumber: "8.1.1", title: "Motion Along a Straight Line" },
        { pointNumber: "8.1.2", title: "Uniform and Non-Uniform Motion" },
      ]},
      { pointNumber: "8.2", title: "Equations of Motion", subtopics: [
        { pointNumber: "8.2.1", title: "First and Second Equations" },
        { pointNumber: "8.2.2", title: "Third Equation and Graphical Method" },
      ]},
      { pointNumber: "8.3", title: "Uniform Circular Motion", subtopics: [
        { pointNumber: "8.3.1", title: "Centripetal Acceleration" },
        { pointNumber: "8.3.2", title: "Examples of Circular Motion" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Force and Laws of Motion",
    topics: [
      { pointNumber: "9.1", title: "Balanced and Unbalanced Forces", subtopics: [
        { pointNumber: "9.1.1", title: "Effect of Forces on Motion" },
        { pointNumber: "9.1.2", title: "First Law of Motion (Inertia)" },
      ]},
      { pointNumber: "9.2", title: "Second Law of Motion", subtopics: [
        { pointNumber: "9.2.1", title: "Momentum" },
        { pointNumber: "9.2.2", title: "F = ma and Its Applications" },
      ]},
      { pointNumber: "9.3", title: "Third Law of Motion", subtopics: [
        { pointNumber: "9.3.1", title: "Action and Reaction" },
        { pointNumber: "9.3.2", title: "Conservation of Momentum" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Gravitation",
    topics: [
      { pointNumber: "10.1", title: "Universal Law of Gravitation", subtopics: [
        { pointNumber: "10.1.1", title: "Gravitational Force" },
        { pointNumber: "10.1.2", title: "Free Fall and g" },
      ]},
      { pointNumber: "10.2", title: "Mass and Weight", subtopics: [
        { pointNumber: "10.2.1", title: "Difference Between Mass and Weight" },
        { pointNumber: "10.2.2", title: "Weight on Moon" },
      ]},
      { pointNumber: "10.3", title: "Thrust and Pressure", subtopics: [
        { pointNumber: "10.3.1", title: "Archimedes Principle" },
        { pointNumber: "10.3.2", title: "Buoyancy and Relative Density" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Work and Energy",
    topics: [
      { pointNumber: "11.1", title: "Work Done by a Force", subtopics: [
        { pointNumber: "11.1.1", title: "Conditions for Work" },
        { pointNumber: "11.1.2", title: "Positive and Negative Work" },
      ]},
      { pointNumber: "11.2", title: "Kinetic and Potential Energy", subtopics: [
        { pointNumber: "11.2.1", title: "Expression for KE" },
        { pointNumber: "11.2.2", title: "Gravitational PE" },
      ]},
      { pointNumber: "11.3", title: "Power and Conservation of Energy", subtopics: [
        { pointNumber: "11.3.1", title: "Definition and Units of Power" },
        { pointNumber: "11.3.2", title: "Law of Conservation of Energy" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Sound",
    topics: [
      { pointNumber: "12.1", title: "Production and Propagation of Sound", subtopics: [
        { pointNumber: "12.1.1", title: "Sound Needs a Medium" },
        { pointNumber: "12.1.2", title: "Speed of Sound in Different Media" },
      ]},
      { pointNumber: "12.2", title: "Reflection of Sound", subtopics: [
        { pointNumber: "12.2.1", title: "Echo and Reverberation" },
        { pointNumber: "12.2.2", title: "Uses of Multiple Reflection" },
      ]},
      { pointNumber: "12.3", title: "Range of Hearing", subtopics: [
        { pointNumber: "12.3.1", title: "Ultrasound and Its Applications" },
        { pointNumber: "12.3.2", title: "SONAR" },
      ]},
    ],
  },

  // ─── Science Class 9 – Chemistry chapters ────────────────────────────────────
  {
    bookName: "Science Class 9", chapterName: "Matter in Our Surroundings",
    topics: [
      { pointNumber: "1.1", title: "Physical Nature of Matter", subtopics: [
        { pointNumber: "1.1.1", title: "Matter Is Made Up of Particles" },
        { pointNumber: "1.1.2", title: "How Small Are These Particles?" },
      ]},
      { pointNumber: "1.2", title: "States of Matter", subtopics: [
        { pointNumber: "1.2.1", title: "Solid, Liquid and Gas" },
        { pointNumber: "1.2.2", title: "Plasma and Bose-Einstein Condensate" },
      ]},
      { pointNumber: "1.3", title: "Interconversion of States", subtopics: [
        { pointNumber: "1.3.1", title: "Effect of Temperature and Pressure" },
        { pointNumber: "1.3.2", title: "Evaporation and Its Effects" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Is Matter Around Us Pure",
    topics: [
      { pointNumber: "2.1", title: "Mixtures", subtopics: [
        { pointNumber: "2.1.1", title: "Homogeneous and Heterogeneous Mixtures" },
        { pointNumber: "2.1.2", title: "Colloids and Suspensions" },
      ]},
      { pointNumber: "2.2", title: "Separating Components of a Mixture", subtopics: [
        { pointNumber: "2.2.1", title: "Evaporation, Centrifugation and Distillation" },
        { pointNumber: "2.2.2", title: "Chromatography" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Atoms and Molecules",
    topics: [
      { pointNumber: "3.1", title: "Laws of Chemical Combination", subtopics: [
        { pointNumber: "3.1.1", title: "Law of Conservation of Mass" },
        { pointNumber: "3.1.2", title: "Law of Constant Proportion" },
      ]},
      { pointNumber: "3.2", title: "Atoms and Molecules", subtopics: [
        { pointNumber: "3.2.1", title: "Atomic Mass and Molecular Mass" },
        { pointNumber: "3.2.2", title: "Mole Concept" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Structure of the Atom",
    topics: [
      { pointNumber: "4.1", title: "Charged Particles in Matter", subtopics: [
        { pointNumber: "4.1.1", title: "Discovery of Electron and Proton" },
        { pointNumber: "4.1.2", title: "Discovery of Neutron" },
      ]},
      { pointNumber: "4.2", title: "Atomic Models", subtopics: [
        { pointNumber: "4.2.1", title: "Thomson's and Rutherford's Model" },
        { pointNumber: "4.2.2", title: "Bohr's Model" },
      ]},
      { pointNumber: "4.3", title: "Valency and Isotopes", subtopics: [
        { pointNumber: "4.3.1", title: "Electronic Configuration and Valency" },
        { pointNumber: "4.3.2", title: "Isotopes and Isobars" },
      ]},
    ],
  },

  // ─── Science Class 9 – Biology chapters ──────────────────────────────────────
  {
    bookName: "Science Class 9", chapterName: "The Fundamental Unit of Life",
    topics: [
      { pointNumber: "5.1", title: "Cell Theory and Discovery", subtopics: [
        { pointNumber: "5.1.1", title: "Who Discovered the Cell?" },
        { pointNumber: "5.1.2", title: "Cell Theory" },
      ]},
      { pointNumber: "5.2", title: "Prokaryotic and Eukaryotic Cells", subtopics: [
        { pointNumber: "5.2.1", title: "Differences Between Prokaryotes and Eukaryotes" },
        { pointNumber: "5.2.2", title: "Plant and Animal Cells" },
      ]},
      { pointNumber: "5.3", title: "Cell Organelles", subtopics: [
        { pointNumber: "5.3.1", title: "Nucleus, Mitochondria and Plastids" },
        { pointNumber: "5.3.2", title: "Vacuole, Endoplasmic Reticulum and Golgi Apparatus" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Tissues",
    topics: [
      { pointNumber: "6.1", title: "Plant Tissues", subtopics: [
        { pointNumber: "6.1.1", title: "Meristematic Tissue" },
        { pointNumber: "6.1.2", title: "Permanent Tissue" },
      ]},
      { pointNumber: "6.2", title: "Animal Tissues", subtopics: [
        { pointNumber: "6.2.1", title: "Epithelial and Connective Tissue" },
        { pointNumber: "6.2.2", title: "Muscular and Nervous Tissue" },
      ]},
    ],
  },
  {
    bookName: "Science Class 9", chapterName: "Why Do We Fall Ill",
    topics: [
      { pointNumber: "13.1", title: "Health and Disease", subtopics: [
        { pointNumber: "13.1.1", title: "Personal and Community Health" },
        { pointNumber: "13.1.2", title: "Distinctions Between Healthy and Disease-free" },
      ]},
      { pointNumber: "13.2", title: "Infectious Diseases", subtopics: [
        { pointNumber: "13.2.1", title: "Causes of Infectious Diseases" },
        { pointNumber: "13.2.2", title: "Means of Spread" },
      ]},
    ],
  },

  // ─── Science Class 10 – Physics chapters ─────────────────────────────────────
  {
    bookName: "Science Class 10", chapterName: "Light – Reflection and Refraction",
    topics: [
      { pointNumber: "10.1", title: "Reflection of Light", subtopics: [
        { pointNumber: "10.1.1", title: "Laws of Reflection" },
        { pointNumber: "10.1.2", title: "Image Formation by Spherical Mirrors" },
      ]},
      { pointNumber: "10.2", title: "Refraction of Light", subtopics: [
        { pointNumber: "10.2.1", title: "Laws of Refraction and Snell's Law" },
        { pointNumber: "10.2.2", title: "Refractive Index" },
      ]},
      { pointNumber: "10.3", title: "Refraction Through a Lens", subtopics: [
        { pointNumber: "10.3.1", title: "Image Formation by Lenses" },
        { pointNumber: "10.3.2", title: "Lens Formula and Magnification" },
      ]},
    ],
  },
  {
    bookName: "Science Class 10", chapterName: "Electricity",
    topics: [
      { pointNumber: "12.1", title: "Electric Current and Circuit", subtopics: [
        { pointNumber: "12.1.1", title: "Electric Potential Difference" },
        { pointNumber: "12.1.2", title: "Ohm's Law" },
      ]},
      { pointNumber: "12.2", title: "Resistance and Resistivity", subtopics: [
        { pointNumber: "12.2.1", title: "Factors Affecting Resistance" },
        { pointNumber: "12.2.2", title: "Resistors in Series and Parallel" },
      ]},
      { pointNumber: "12.3", title: "Heating Effect of Electric Current", subtopics: [
        { pointNumber: "12.3.1", title: "Joule's Law of Heating" },
        { pointNumber: "12.3.2", title: "Practical Applications" },
      ]},
    ],
  },
  {
    bookName: "Science Class 10", chapterName: "Magnetic Effects of Electric Current",
    topics: [
      { pointNumber: "13.1", title: "Magnetic Field and Field Lines", subtopics: [
        { pointNumber: "13.1.1", title: "Magnetic Field Due to Current" },
        { pointNumber: "13.1.2", title: "Solenoid and Electromagnet" },
      ]},
      { pointNumber: "13.2", title: "Force on Current-Carrying Conductor", subtopics: [
        { pointNumber: "13.2.1", title: "Fleming's Left-Hand Rule" },
        { pointNumber: "13.2.2", title: "Electric Motor" },
      ]},
      { pointNumber: "13.3", title: "Electromagnetic Induction", subtopics: [
        { pointNumber: "13.3.1", title: "Fleming's Right-Hand Rule" },
        { pointNumber: "13.3.2", title: "Electric Generator and Domestic Circuits" },
      ]},
    ],
  },

  // ─── Science Class 10 – Chemistry chapters ───────────────────────────────────
  {
    bookName: "Science Class 10", chapterName: "Chemical Reactions and Equations",
    topics: [
      { pointNumber: "1.1", title: "Chemical Equations", subtopics: [
        { pointNumber: "1.1.1", title: "Writing a Chemical Equation" },
        { pointNumber: "1.1.2", title: "Balancing a Chemical Equation" },
      ]},
      { pointNumber: "1.2", title: "Types of Chemical Reactions", subtopics: [
        { pointNumber: "1.2.1", title: "Combination and Decomposition Reactions" },
        { pointNumber: "1.2.2", title: "Displacement and Double Displacement" },
      ]},
    ],
  },
  {
    bookName: "Science Class 10", chapterName: "Acids, Bases and Salts",
    topics: [
      { pointNumber: "2.1", title: "Understanding Acids and Bases", subtopics: [
        { pointNumber: "2.1.1", title: "Chemical Properties of Acids" },
        { pointNumber: "2.1.2", title: "Chemical Properties of Bases" },
      ]},
      { pointNumber: "2.2", title: "pH Scale and Salts", subtopics: [
        { pointNumber: "2.2.1", title: "Importance of pH in Daily Life" },
        { pointNumber: "2.2.2", title: "Common Salt, Baking Soda and Washing Soda" },
      ]},
    ],
  },
  {
    bookName: "Science Class 10", chapterName: "Metals and Non-metals",
    topics: [
      { pointNumber: "3.1", title: "Physical Properties of Metals", subtopics: [
        { pointNumber: "3.1.1", title: "Malleability, Ductility and Conductivity" },
        { pointNumber: "3.1.2", title: "Properties of Non-metals" },
      ]},
      { pointNumber: "3.2", title: "Chemical Properties and Reactivity Series", subtopics: [
        { pointNumber: "3.2.1", title: "Reactions of Metals with Water and Acids" },
        { pointNumber: "3.2.2", title: "Ionic Bonding and Corrosion" },
      ]},
    ],
  },

  // ─── Science Class 10 – Biology chapters ─────────────────────────────────────
  {
    bookName: "Science Class 10", chapterName: "Life Processes",
    topics: [
      { pointNumber: "6.1", title: "Nutrition", subtopics: [
        { pointNumber: "6.1.1", title: "Autotrophic Nutrition" },
        { pointNumber: "6.1.2", title: "Heterotrophic Nutrition" },
      ]},
      { pointNumber: "6.2", title: "Respiration", subtopics: [
        { pointNumber: "6.2.1", title: "Aerobic and Anaerobic Respiration" },
        { pointNumber: "6.2.2", title: "Human Respiratory System" },
      ]},
      { pointNumber: "6.3", title: "Transportation and Excretion", subtopics: [
        { pointNumber: "6.3.1", title: "Blood and Lymph Transport" },
        { pointNumber: "6.3.2", title: "Excretory System in Humans" },
      ]},
    ],
  },
  {
    bookName: "Science Class 10", chapterName: "Heredity and Evolution",
    topics: [
      { pointNumber: "9.1", title: "Heredity", subtopics: [
        { pointNumber: "9.1.1", title: "Accumulation of Variation During Reproduction" },
        { pointNumber: "9.1.2", title: "Mendel's Contributions" },
      ]},
      { pointNumber: "9.2", title: "Evolution", subtopics: [
        { pointNumber: "9.2.1", title: "Darwin's Theory of Natural Selection" },
        { pointNumber: "9.2.2", title: "Speciation and Evolutionary Relationships" },
      ]},
    ],
  },

  // ─── Mathematics Class 9 (CBSE) ──────────────────────────────────────────────
  {
    bookName: "Mathematics Class 9", chapterName: "Number Systems",
    topics: [
      { pointNumber: "1.1", title: "Irrational Numbers", subtopics: [
        { pointNumber: "1.1.1", title: "Real Numbers and Their Decimal Expansions" },
        { pointNumber: "1.1.2", title: "Representing Real Numbers on Number Line" },
      ]},
      { pointNumber: "1.2", title: "Operations on Real Numbers", subtopics: [
        { pointNumber: "1.2.1", title: "Identities for Irrational Numbers" },
        { pointNumber: "1.2.2", title: "Laws of Exponents for Real Numbers" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 9", chapterName: "Triangles",
    topics: [
      { pointNumber: "7.1", title: "Congruence of Triangles", subtopics: [
        { pointNumber: "7.1.1", title: "Criteria for Congruence" },
        { pointNumber: "7.1.2", title: "SAS, ASA and AAS Congruence" },
      ]},
      { pointNumber: "7.2", title: "Inequalities in a Triangle", subtopics: [
        { pointNumber: "7.2.1", title: "Angle and Side Inequalities" },
        { pointNumber: "7.2.2", title: "Triangle Inequality" },
      ]},
    ],
  },

  // ─── Mathematics Class 10 (CBSE) ─────────────────────────────────────────────
  {
    bookName: "Mathematics Class 10", chapterName: "Real Numbers",
    topics: [
      { pointNumber: "1.1", title: "Euclid's Division Lemma", subtopics: [
        { pointNumber: "1.1.1", title: "Euclid's Division Algorithm" },
        { pointNumber: "1.1.2", title: "Finding HCF Using Euclid's Algorithm" },
      ]},
      { pointNumber: "1.2", title: "The Fundamental Theorem of Arithmetic", subtopics: [
        { pointNumber: "1.2.1", title: "HCF and LCM Using Prime Factorisation" },
        { pointNumber: "1.2.2", title: "Irrational and Rational Numbers" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 10", chapterName: "Quadratic Equations",
    topics: [
      { pointNumber: "4.1", title: "Introduction to Quadratic Equations", subtopics: [
        { pointNumber: "4.1.1", title: "Standard Form of a Quadratic Equation" },
        { pointNumber: "4.1.2", title: "Solution by Factorisation" },
      ]},
      { pointNumber: "4.2", title: "Nature of Roots", subtopics: [
        { pointNumber: "4.2.1", title: "Quadratic Formula" },
        { pointNumber: "4.2.2", title: "Discriminant and Nature of Roots" },
      ]},
    ],
  },

  // ─── Physics Class 11 (CBSE) ─────────────────────────────────────────────────
  {
    bookName: "Physics Class 11", chapterName: "Units and Measurements",
    topics: [
      { pointNumber: "2.1", title: "The International System of Units", subtopics: [
        { pointNumber: "2.1.1", title: "Base and Derived Units" },
        { pointNumber: "2.1.2", title: "Measurement of Length, Mass and Time" },
      ]},
      { pointNumber: "2.2", title: "Significant Figures and Errors", subtopics: [
        { pointNumber: "2.2.1", title: "Absolute and Relative Errors" },
        { pointNumber: "2.2.2", title: "Dimensional Analysis" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 11", chapterName: "Motion in a Straight Line",
    topics: [
      { pointNumber: "3.1", title: "Position, Path Length and Displacement", subtopics: [
        { pointNumber: "3.1.1", title: "Uniform and Non-Uniform Motion" },
        { pointNumber: "3.1.2", title: "Instantaneous Velocity and Speed" },
      ]},
      { pointNumber: "3.2", title: "Acceleration and Kinematic Equations", subtopics: [
        { pointNumber: "3.2.1", title: "Equations of Uniformly Accelerated Motion" },
        { pointNumber: "3.2.2", title: "Relative Velocity" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 11", chapterName: "Laws of Motion",
    topics: [
      { pointNumber: "5.1", title: "Aristotle's Fallacy and Newton's First Law", subtopics: [
        { pointNumber: "5.1.1", title: "Inertia and Mass" },
        { pointNumber: "5.1.2", title: "Newton's First Law of Motion" },
      ]},
      { pointNumber: "5.2", title: "Newton's Second and Third Laws", subtopics: [
        { pointNumber: "5.2.1", title: "Impulse and Momentum" },
        { pointNumber: "5.2.2", title: "Conservation of Momentum" },
      ]},
      { pointNumber: "5.3", title: "Friction", subtopics: [
        { pointNumber: "5.3.1", title: "Static and Kinetic Friction" },
        { pointNumber: "5.3.2", title: "Circular Motion and Centripetal Force" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 11", chapterName: "Work, Energy and Power",
    topics: [
      { pointNumber: "6.1", title: "Work Done by a Constant Force", subtopics: [
        { pointNumber: "6.1.1", title: "Work-Energy Theorem" },
        { pointNumber: "6.1.2", title: "Potential Energy" },
      ]},
      { pointNumber: "6.2", title: "Power and Collisions", subtopics: [
        { pointNumber: "6.2.1", title: "Elastic and Inelastic Collisions" },
        { pointNumber: "6.2.2", title: "Conservation of Mechanical Energy" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 11", chapterName: "Oscillations",
    topics: [
      { pointNumber: "14.1", title: "Periodic and Oscillatory Motion", subtopics: [
        { pointNumber: "14.1.1", title: "Displacement in SHM" },
        { pointNumber: "14.1.2", title: "Velocity and Acceleration in SHM" },
      ]},
      { pointNumber: "14.2", title: "Energy in SHM and Pendulum", subtopics: [
        { pointNumber: "14.2.1", title: "Simple Pendulum" },
        { pointNumber: "14.2.2", title: "Damped and Forced Oscillations" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 11", chapterName: "Waves",
    topics: [
      { pointNumber: "15.1", title: "Transverse and Longitudinal Waves", subtopics: [
        { pointNumber: "15.1.1", title: "Displacement Relation for a Wave" },
        { pointNumber: "15.1.2", title: "Speed of a Travelling Wave" },
      ]},
      { pointNumber: "15.2", title: "Reflection and Superposition", subtopics: [
        { pointNumber: "15.2.1", title: "Standing Waves and Normal Modes" },
        { pointNumber: "15.2.2", title: "Beats and Doppler Effect" },
      ]},
    ],
  },

  // ─── Physics Class 12 (CBSE) ─────────────────────────────────────────────────
  {
    bookName: "Physics Class 12", chapterName: "Electric Charges and Fields",
    topics: [
      { pointNumber: "1.1", title: "Electric Charge and Coulomb's Law", subtopics: [
        { pointNumber: "1.1.1", title: "Basic Properties of Electric Charge" },
        { pointNumber: "1.1.2", title: "Coulomb's Law" },
      ]},
      { pointNumber: "1.2", title: "Electric Field and Flux", subtopics: [
        { pointNumber: "1.2.1", title: "Electric Field Lines" },
        { pointNumber: "1.2.2", title: "Gauss's Law" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 12", chapterName: "Current Electricity",
    topics: [
      { pointNumber: "3.1", title: "Electric Current and Drift Velocity", subtopics: [
        { pointNumber: "3.1.1", title: "Ohm's Law and Resistance" },
        { pointNumber: "3.1.2", title: "Resistivity and Its Temperature Dependence" },
      ]},
      { pointNumber: "3.2", title: "Cells and Kirchhoff's Laws", subtopics: [
        { pointNumber: "3.2.1", title: "EMF and Internal Resistance" },
        { pointNumber: "3.2.2", title: "Wheatstone Bridge and Potentiometer" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 12", chapterName: "Ray Optics and Optical Instruments",
    topics: [
      { pointNumber: "9.1", title: "Reflection and Refraction", subtopics: [
        { pointNumber: "9.1.1", title: "Mirror Formula and Lens Formula" },
        { pointNumber: "9.1.2", title: "Total Internal Reflection" },
      ]},
      { pointNumber: "9.2", title: "Optical Instruments", subtopics: [
        { pointNumber: "9.2.1", title: "Microscope and Telescope" },
        { pointNumber: "9.2.2", title: "Eye and Its Defects" },
      ]},
    ],
  },
  {
    bookName: "Physics Class 12", chapterName: "Atoms",
    topics: [
      { pointNumber: "12.1", title: "Alpha-Particle Scattering and Rutherford's Model", subtopics: [
        { pointNumber: "12.1.1", title: "Bohr's Model of Hydrogen Atom" },
        { pointNumber: "12.1.2", title: "Energy Levels and Spectra" },
      ]},
      { pointNumber: "12.2", title: "Hydrogen Spectrum", subtopics: [
        { pointNumber: "12.2.1", title: "Line Spectra of Hydrogen" },
        { pointNumber: "12.2.2", title: "de Broglie's Explanation" },
      ]},
    ],
  },

  // ─── Chemistry Class 11 (CBSE) ───────────────────────────────────────────────
  {
    bookName: "Chemistry Class 11", chapterName: "Some Basic Concepts of Chemistry",
    topics: [
      { pointNumber: "1.1", title: "Importance and Scope of Chemistry", subtopics: [
        { pointNumber: "1.1.1", title: "Nature of Matter" },
        { pointNumber: "1.1.2", title: "Properties of Matter" },
      ]},
      { pointNumber: "1.2", title: "Mole Concept and Stoichiometry", subtopics: [
        { pointNumber: "1.2.1", title: "Atomic and Molecular Masses" },
        { pointNumber: "1.2.2", title: "Limiting Reagent and Percent Yield" },
      ]},
    ],
  },
  {
    bookName: "Chemistry Class 11", chapterName: "Structure of Atom",
    topics: [
      { pointNumber: "2.1", title: "Discovery of Sub-Atomic Particles", subtopics: [
        { pointNumber: "2.1.1", title: "Thomson and Rutherford Models" },
        { pointNumber: "2.1.2", title: "Atomic Number and Mass Number" },
      ]},
      { pointNumber: "2.2", title: "Quantum Mechanical Model", subtopics: [
        { pointNumber: "2.2.1", title: "Quantum Numbers and Orbitals" },
        { pointNumber: "2.2.2", title: "Electronic Configuration and Aufbau Principle" },
      ]},
    ],
  },
  {
    bookName: "Chemistry Class 11", chapterName: "Thermodynamics",
    topics: [
      { pointNumber: "5.1", title: "Thermodynamic Terms and State Functions", subtopics: [
        { pointNumber: "5.1.1", title: "Internal Energy and Enthalpy" },
        { pointNumber: "5.1.2", title: "First Law of Thermodynamics" },
      ]},
      { pointNumber: "5.2", title: "Entropy and Gibbs Energy", subtopics: [
        { pointNumber: "5.2.1", title: "Second and Third Laws of Thermodynamics" },
        { pointNumber: "5.2.2", title: "Gibbs Free Energy and Spontaneity" },
      ]},
    ],
  },

  // ─── Chemistry Class 12 (CBSE) ───────────────────────────────────────────────
  {
    bookName: "Chemistry Class 12", chapterName: "The Solid State",
    topics: [
      { pointNumber: "1.1", title: "General Characteristics of Solid State", subtopics: [
        { pointNumber: "1.1.1", title: "Crystalline and Amorphous Solids" },
        { pointNumber: "1.1.2", title: "Classification of Crystalline Solids" },
      ]},
      { pointNumber: "1.2", title: "Crystal Lattices and Unit Cells", subtopics: [
        { pointNumber: "1.2.1", title: "Packing Efficiency and Density" },
        { pointNumber: "1.2.2", title: "Imperfections in Solids" },
      ]},
    ],
  },
  {
    bookName: "Chemistry Class 12", chapterName: "Electrochemistry",
    topics: [
      { pointNumber: "3.1", title: "Electrochemical Cells and EMF", subtopics: [
        { pointNumber: "3.1.1", title: "Galvanic Cells and Standard Electrode Potential" },
        { pointNumber: "3.1.2", title: "Nernst Equation" },
      ]},
      { pointNumber: "3.2", title: "Electrolytic Cells and Conductance", subtopics: [
        { pointNumber: "3.2.1", title: "Faraday's Laws of Electrolysis" },
        { pointNumber: "3.2.2", title: "Batteries and Fuel Cells" },
      ]},
    ],
  },
  {
    bookName: "Chemistry Class 12", chapterName: "Chemical Kinetics",
    topics: [
      { pointNumber: "4.1", title: "Rate of a Chemical Reaction", subtopics: [
        { pointNumber: "4.1.1", title: "Factors Affecting Reaction Rate" },
        { pointNumber: "4.1.2", title: "Order and Molecularity of Reactions" },
      ]},
      { pointNumber: "4.2", title: "Integrated Rate Equations and Activation Energy", subtopics: [
        { pointNumber: "4.2.1", title: "First Order Reactions and Half-Life" },
        { pointNumber: "4.2.2", title: "Arrhenius Equation and Catalysis" },
      ]},
    ],
  },

  // ─── Biology Class 11 (CBSE) ─────────────────────────────────────────────────
  {
    bookName: "Biology Class 11", chapterName: "The Living World",
    topics: [
      { pointNumber: "1.1", title: "What is Living?", subtopics: [
        { pointNumber: "1.1.1", title: "Diversity in Living Organisms" },
        { pointNumber: "1.1.2", title: "Taxonomic Categories" },
      ]},
      { pointNumber: "1.2", title: "Nomenclature and Classification", subtopics: [
        { pointNumber: "1.2.1", title: "Binomial Nomenclature" },
        { pointNumber: "1.2.2", title: "Tools for Study of Taxonomy" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 11", chapterName: "Cell: The Unit of Life",
    topics: [
      { pointNumber: "8.1", title: "Cell Theory and Cell Types", subtopics: [
        { pointNumber: "8.1.1", title: "Prokaryotic and Eukaryotic Cells" },
        { pointNumber: "8.1.2", title: "Plant and Animal Cell Differences" },
      ]},
      { pointNumber: "8.2", title: "Cell Organelles", subtopics: [
        { pointNumber: "8.2.1", title: "Mitochondria, Plastids and Ribosomes" },
        { pointNumber: "8.2.2", title: "Endomembrane System" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 11", chapterName: "Photosynthesis in Higher Plants",
    topics: [
      { pointNumber: "11.1", title: "Early Experiments and Site of Photosynthesis", subtopics: [
        { pointNumber: "11.1.1", title: "Chloroplast Structure" },
        { pointNumber: "11.1.2", title: "Pigments Involved in Photosynthesis" },
      ]},
      { pointNumber: "11.2", title: "Light Reactions and Calvin Cycle", subtopics: [
        { pointNumber: "11.2.1", title: "Electron Transport Chain and ATP Synthesis" },
        { pointNumber: "11.2.2", title: "C3 and C4 Pathways" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 11", chapterName: "Neural Control and Coordination",
    topics: [
      { pointNumber: "18.1", title: "Neural System", subtopics: [
        { pointNumber: "18.1.1", title: "Neuron as Structural and Functional Unit" },
        { pointNumber: "18.1.2", title: "Generation and Conduction of Nerve Impulse" },
      ]},
      { pointNumber: "18.2", title: "Central and Peripheral Nervous System", subtopics: [
        { pointNumber: "18.2.1", title: "Structure of Brain" },
        { pointNumber: "18.2.2", title: "Reflex Action and Sense Organs" },
      ]},
    ],
  },

  // ─── Biology Class 12 (CBSE) ─────────────────────────────────────────────────
  {
    bookName: "Biology Class 12", chapterName: "Principles of Inheritance and Variation",
    topics: [
      { pointNumber: "4.1", title: "Mendel's Laws of Inheritance", subtopics: [
        { pointNumber: "4.1.1", title: "Law of Dominance and Segregation" },
        { pointNumber: "4.1.2", title: "Law of Independent Assortment" },
      ]},
      { pointNumber: "4.2", title: "Chromosomal Theory and Linkage", subtopics: [
        { pointNumber: "4.2.1", title: "Linkage and Recombination" },
        { pointNumber: "4.2.2", title: "Sex Determination" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 12", chapterName: "Molecular Basis of Inheritance",
    topics: [
      { pointNumber: "5.1", title: "DNA as Genetic Material", subtopics: [
        { pointNumber: "5.1.1", title: "Structure of DNA" },
        { pointNumber: "5.1.2", title: "DNA Packaging and Replication" },
      ]},
      { pointNumber: "5.2", title: "Transcription and Translation", subtopics: [
        { pointNumber: "5.2.1", title: "Transcription and Genetic Code" },
        { pointNumber: "5.2.2", title: "Translation and Gene Regulation" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 12", chapterName: "Evolution",
    topics: [
      { pointNumber: "6.1", title: "Origin of Life", subtopics: [
        { pointNumber: "6.1.1", title: "Chemical Evolution" },
        { pointNumber: "6.1.2", title: "Evidence for Evolution" },
      ]},
      { pointNumber: "6.2", title: "Mechanism of Evolution", subtopics: [
        { pointNumber: "6.2.1", title: "Darwinism and Neo-Darwinism" },
        { pointNumber: "6.2.2", title: "Hardy-Weinberg Equilibrium" },
      ]},
    ],
  },
  {
    bookName: "Biology Class 12", chapterName: "Ecosystem",
    topics: [
      { pointNumber: "12.1", title: "Ecosystem: Structure and Function", subtopics: [
        { pointNumber: "12.1.1", title: "Productivity and Decomposition" },
        { pointNumber: "12.1.2", title: "Energy Flow" },
      ]},
      { pointNumber: "12.2", title: "Ecological Pyramids and Succession", subtopics: [
        { pointNumber: "12.2.1", title: "Nutrient Cycling" },
        { pointNumber: "12.2.2", title: "Ecosystem Services" },
      ]},
    ],
  },

  // ─── Mathematics Class 11 (CBSE) ─────────────────────────────────────────────
  {
    bookName: "Mathematics Class 11", chapterName: "Sets",
    topics: [
      { pointNumber: "1.1", title: "Sets and Their Representations", subtopics: [
        { pointNumber: "1.1.1", title: "Empty Set, Finite and Infinite Sets" },
        { pointNumber: "1.1.2", title: "Subsets and Power Set" },
      ]},
      { pointNumber: "1.2", title: "Operations on Sets", subtopics: [
        { pointNumber: "1.2.1", title: "Union, Intersection and Difference" },
        { pointNumber: "1.2.2", title: "Venn Diagrams and Practical Problems" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 11", chapterName: "Trigonometric Functions",
    topics: [
      { pointNumber: "3.1", title: "Angles and Trigonometric Functions", subtopics: [
        { pointNumber: "3.1.1", title: "Degree and Radian Measure" },
        { pointNumber: "3.1.2", title: "Trigonometric Functions of Any Angle" },
      ]},
      { pointNumber: "3.2", title: "Trigonometric Identities and Equations", subtopics: [
        { pointNumber: "3.2.1", title: "Sum and Difference Identities" },
        { pointNumber: "3.2.2", title: "General Solution of Trigonometric Equations" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 11", chapterName: "Limits and Derivatives",
    topics: [
      { pointNumber: "12.1", title: "Intuitive Idea of Derivatives", subtopics: [
        { pointNumber: "12.1.1", title: "Limits of Functions" },
        { pointNumber: "12.1.2", title: "Algebra of Limits" },
      ]},
      { pointNumber: "12.2", title: "Derivatives", subtopics: [
        { pointNumber: "12.2.1", title: "Derivative of a Function" },
        { pointNumber: "12.2.2", title: "Algebra of Derivatives" },
      ]},
    ],
  },

  // ─── Mathematics Class 12 (CBSE) ─────────────────────────────────────────────
  {
    bookName: "Mathematics Class 12", chapterName: "Matrices",
    topics: [
      { pointNumber: "3.1", title: "Matrix and Types of Matrices", subtopics: [
        { pointNumber: "3.1.1", title: "Operations on Matrices" },
        { pointNumber: "3.1.2", title: "Transpose and Symmetric Matrices" },
      ]},
      { pointNumber: "3.2", title: "Elementary Operations and Invertible Matrices", subtopics: [
        { pointNumber: "3.2.1", title: "Inverse of a Matrix" },
        { pointNumber: "3.2.2", title: "Solving Systems of Equations Using Matrices" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 12", chapterName: "Integrals",
    topics: [
      { pointNumber: "7.1", title: "Integration as Inverse of Differentiation", subtopics: [
        { pointNumber: "7.1.1", title: "Integration by Substitution" },
        { pointNumber: "7.1.2", title: "Integration by Partial Fractions" },
      ]},
      { pointNumber: "7.2", title: "Definite Integrals", subtopics: [
        { pointNumber: "7.2.1", title: "Definite Integral as Limit of a Sum" },
        { pointNumber: "7.2.2", title: "Fundamental Theorem of Calculus" },
      ]},
    ],
  },
  {
    bookName: "Mathematics Class 12", chapterName: "Probability",
    topics: [
      { pointNumber: "13.1", title: "Conditional Probability", subtopics: [
        { pointNumber: "13.1.1", title: "Multiplication Theorem on Probability" },
        { pointNumber: "13.1.2", title: "Independent Events" },
      ]},
      { pointNumber: "13.2", title: "Bayes' Theorem and Random Variables", subtopics: [
        { pointNumber: "13.2.1", title: "Bayes' Theorem" },
        { pointNumber: "13.2.2", title: "Binomial Distribution" },
      ]},
    ],
  },
];

async function main() {
  const allChapters = await prisma.chapterMaster.findMany({ include: { book: true } });
  const chapterMap = new Map<string, string>();
  for (const ch of allChapters) {
    chapterMap.set(`${ch.book.name}|||${ch.name}`, ch.id);
  }

  await prisma.topicMaster.deleteMany();
  console.log("Cleared existing topics and subtopics");

  let topicCount = 0;
  let subtopicCount = 0;
  let skipped = 0;

  for (const entry of DATA) {
    const key = `${entry.bookName}|||${entry.chapterName}`;
    const chapterId = chapterMap.get(key);
    if (!chapterId) {
      console.warn(`  [SKIP] Chapter not found: ${key}`);
      skipped++;
      continue;
    }

    for (let ti = 0; ti < entry.topics.length; ti++) {
      const t = entry.topics[ti];
      const topic = await prisma.topicMaster.create({
        data: {
          chapterId,
          pointNumber: t.pointNumber,
          title: t.title,
          sortOrder: ti + 1,
          status: "active",
        },
      });
      topicCount++;

      if (t.subtopics) {
        for (let si = 0; si < t.subtopics.length; si++) {
          const s = t.subtopics[si];
          await prisma.subtopicMaster.create({
            data: {
              topicId: topic.id,
              pointNumber: s.pointNumber,
              title: s.title,
              sortOrder: si + 1,
              status: "active",
            },
          });
          subtopicCount++;
        }
      }
    }
  }

  console.log(`Topics created   : ${topicCount}`);
  console.log(`Subtopics created: ${subtopicCount}`);
  if (skipped > 0) console.warn(`Chapters skipped : ${skipped}`);
}

main()
  .catch(console.error)
  .finally(() => prisma.$disconnect());
