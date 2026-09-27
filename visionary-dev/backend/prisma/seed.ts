import { PrismaClient } from "@prisma/client";

const prisma = new PrismaClient();

// ─── Icon map ─────────────────────────────────────────────────────────────────

const ICON: Record<string, string> = {
  "Mathematics":        "/assets/Mathematices.png",
  "Science":            "/assets/Science-icon.png",
  "Physics":            "/assets/physics.png",
  "Chemistry":          "/assets/chemistry.png",
  "Biology":            "/assets/biology.png",
  "Social Studies":     "/assets/social-studies.png",
  "Social Science":     "/assets/social-studies.png",
  "English":            "/assets/english.png",
  "Hindi":              "/assets/hindi.png",
};

function icon(name: string): string | null {
  return ICON[name] ?? null;
}

// ─── Boards ───────────────────────────────────────────────────────────────────

const boards = [
  { name: "CBSE", sortOrder: 0 },
  { name: "ICSE", sortOrder: 0 },
  { name: "Andhra Pradesh Board of Secondary Education", sortOrder: 1 },
  { name: "Assam Higher Secondary Education Council", sortOrder: 2 },
  { name: "Bihar School Examination Board", sortOrder: 3 },
  { name: "Chhattisgarh Board of Secondary Education", sortOrder: 4 },
  { name: "Goa Board of Secondary & Higher Secondary Education", sortOrder: 5 },
  { name: "Gujarat Secondary and Higher Secondary Education Board", sortOrder: 6 },
  { name: "Haryana Board of School Education", sortOrder: 7 },
  { name: "Himachal Pradesh Board of School Education", sortOrder: 8 },
  { name: "Jammu & Kashmir State Board of School Education", sortOrder: 9 },
  { name: "Jharkhand Academic Council", sortOrder: 10 },
  { name: "Karnataka Secondary Education Examination Board", sortOrder: 11 },
  { name: "Kerala Board of Public Examinations", sortOrder: 12 },
  { name: "Madhya Pradesh Board of Secondary Education", sortOrder: 13 },
  { name: "Maharashtra State Board of Secondary and Higher Secondary Education", sortOrder: 14 },
  { name: "Manipur Board of Secondary Education", sortOrder: 15 },
  { name: "Meghalaya Board of School Education", sortOrder: 16 },
  { name: "Mizoram Board of School Education", sortOrder: 17 },
  { name: "Nagaland Board of School Education", sortOrder: 18 },
  { name: "Odisha Board of Secondary Education", sortOrder: 19 },
  { name: "Punjab School Education Board", sortOrder: 20 },
  { name: "Rajasthan Board of Secondary Education", sortOrder: 21 },
  { name: "Sikkim Human Resource Development Institute", sortOrder: 22 },
  { name: "Tamil Nadu Board of Secondary Education", sortOrder: 23 },
  { name: "Telangana Board of Secondary Education", sortOrder: 24 },
  { name: "Tripura Board of Secondary Education", sortOrder: 25 },
  { name: "Uttar Pradesh Madhyamik Shiksha Parishad", sortOrder: 26 },
  { name: "Uttarakhand Board of School Education", sortOrder: 27 },
  { name: "West Bengal Board of Secondary Education", sortOrder: 28 },
  { name: "International Baccalaureate (IB)", sortOrder: 29 },
  { name: "Cambridge International (IGCSE)", sortOrder: 30 },
];

// ─── Grades ───────────────────────────────────────────────────────────────────

const grades = [
  { name: "1st",  sortOrder: 1  },
  { name: "2nd",  sortOrder: 2  },
  { name: "3rd",  sortOrder: 3  },
  { name: "4th",  sortOrder: 4  },
  { name: "5th",  sortOrder: 5  },
  { name: "6th",  sortOrder: 6  },
  { name: "7th",  sortOrder: 7  },
  { name: "8th",  sortOrder: 8  },
  { name: "9th",  sortOrder: 9  },
  { name: "10th", sortOrder: 10 },
  { name: "11th", sortOrder: 11 },
  { name: "12th", sortOrder: 12 },
];

// ─── Subject definitions ──────────────────────────────────────────────────────
// boardId = null  → common (shown for any board that has no specific data)
// boardId = cbseId → CBSE NCERT curriculum

type SubjectDef = { name: string; sortOrder: number };

// Common subjects – base curriculum for all non-specific boards
const commonByGrade: Record<string, SubjectDef[]> = {
  "1st":  [
    { name: "Mathematics", sortOrder: 1 },
    { name: "English",     sortOrder: 2 },
    { name: "Hindi",       sortOrder: 3 },
    { name: "Science",     sortOrder: 4 },
  ],
  "2nd":  [
    { name: "Mathematics", sortOrder: 1 },
    { name: "English",     sortOrder: 2 },
    { name: "Hindi",       sortOrder: 3 },
    { name: "Science",     sortOrder: 4 },
  ],
  "3rd":  [
    { name: "Mathematics",  sortOrder: 1 },
    { name: "English",      sortOrder: 2 },
    { name: "Hindi",        sortOrder: 3 },
    { name: "Science",      sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "4th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "5th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "6th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "7th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "8th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "9th":  [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Physics",       sortOrder: 4 },
    { name: "Chemistry",     sortOrder: 5 },
    { name: "Biology",       sortOrder: 6 },
    { name: "Social Science", sortOrder: 7 },
    { name: "Computer Science", sortOrder: 8 },
  ],
  "10th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Physics",       sortOrder: 4 },
    { name: "Chemistry",     sortOrder: 5 },
    { name: "Biology",       sortOrder: 6 },
    { name: "Social Science", sortOrder: 7 },
    { name: "Computer Science", sortOrder: 8 },
  ],
  "11th": [
    { name: "Physics",       sortOrder: 1 },
    { name: "Chemistry",     sortOrder: 2 },
    { name: "Biology",       sortOrder: 3 },
    { name: "Mathematics",   sortOrder: 4 },
    { name: "English",       sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
    { name: "Economics",     sortOrder: 7 },
    { name: "Business Studies", sortOrder: 8 },
    { name: "Accountancy",   sortOrder: 9 },
    { name: "History",       sortOrder: 10 },
    { name: "Political Science", sortOrder: 11 },
    { name: "Geography",     sortOrder: 12 },
    { name: "Physical Education", sortOrder: 13 },
  ],
  "12th": [
    { name: "Physics",       sortOrder: 1 },
    { name: "Chemistry",     sortOrder: 2 },
    { name: "Biology",       sortOrder: 3 },
    { name: "Mathematics",   sortOrder: 4 },
    { name: "English",       sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
    { name: "Economics",     sortOrder: 7 },
    { name: "Business Studies", sortOrder: 8 },
    { name: "Accountancy",   sortOrder: 9 },
    { name: "History",       sortOrder: 10 },
    { name: "Political Science", sortOrder: 11 },
    { name: "Geography",     sortOrder: 12 },
    { name: "Physical Education", sortOrder: 13 },
  ],
};

// CBSE NCERT subjects — used when boardId = cbseId
const cbseByGrade: Record<string, SubjectDef[]> = {
  "1st": [
    { name: "Mathematics",   sortOrder: 1 },  // Math Magic
    { name: "English",       sortOrder: 2 },  // Marigold
    { name: "Hindi",         sortOrder: 3 },  // Rimjhim
    { name: "Science",       sortOrder: 4 },  // EVS / Looking Around
  ],
  "2nd": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
  ],
  "3rd": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "4th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "5th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Studies", sortOrder: 5 },
  ],
  "6th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Science", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "7th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Science", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "8th": [
    { name: "Mathematics",   sortOrder: 1 },
    { name: "English",       sortOrder: 2 },
    { name: "Hindi",         sortOrder: 3 },
    { name: "Science",       sortOrder: 4 },
    { name: "Social Science", sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
  ],
  "9th": [
    { name: "Mathematics",     sortOrder: 1 },
    { name: "English",         sortOrder: 2 },
    { name: "Hindi",           sortOrder: 3 },
    { name: "Physics",         sortOrder: 4 },
    { name: "Chemistry",       sortOrder: 5 },
    { name: "Biology",         sortOrder: 6 },
    { name: "Social Science",  sortOrder: 7 },
    { name: "Computer Science", sortOrder: 8 },
  ],
  "10th": [
    { name: "Mathematics",     sortOrder: 1 },
    { name: "English",         sortOrder: 2 },
    { name: "Hindi",           sortOrder: 3 },
    { name: "Physics",         sortOrder: 4 },
    { name: "Chemistry",       sortOrder: 5 },
    { name: "Biology",         sortOrder: 6 },
    { name: "Social Science",  sortOrder: 7 },
    { name: "Computer Science", sortOrder: 8 },
  ],
  "11th": [
    { name: "Physics",         sortOrder: 1 },
    { name: "Chemistry",       sortOrder: 2 },
    { name: "Biology",         sortOrder: 3 },
    { name: "Mathematics",     sortOrder: 4 },
    { name: "English",         sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
    { name: "Economics",       sortOrder: 7 },
    { name: "Business Studies", sortOrder: 8 },
    { name: "Accountancy",     sortOrder: 9 },
    { name: "History",         sortOrder: 10 },
    { name: "Political Science", sortOrder: 11 },
    { name: "Geography",       sortOrder: 12 },
    { name: "Physical Education", sortOrder: 13 },
  ],
  "12th": [
    { name: "Physics",         sortOrder: 1 },
    { name: "Chemistry",       sortOrder: 2 },
    { name: "Biology",         sortOrder: 3 },
    { name: "Mathematics",     sortOrder: 4 },
    { name: "English",         sortOrder: 5 },
    { name: "Computer Science", sortOrder: 6 },
    { name: "Economics",       sortOrder: 7 },
    { name: "Business Studies", sortOrder: 8 },
    { name: "Accountancy",     sortOrder: 9 },
    { name: "History",         sortOrder: 10 },
    { name: "Political Science", sortOrder: 11 },
    { name: "Geography",       sortOrder: 12 },
    { name: "Physical Education", sortOrder: 13 },
  ],
};

// ─── Main ─────────────────────────────────────────────────────────────────────

async function main() {
  // 1. Boards
  for (const board of boards) {
    await prisma.boardMaster.upsert({
      where: { name: board.name },
      update: {},
      create: board,
    });
  }
  console.log(`✓ Seeded ${boards.length} boards`);

  // 2. Grades
  for (const grade of grades) {
    await prisma.gradeMaster.upsert({
      where: { name: grade.name },
      update: {},
      create: grade,
    });
  }
  console.log(`✓ Seeded ${grades.length} grades`);

  // 3. Subjects — delete all first, then recreate cleanly
  await prisma.subjectMaster.deleteMany();

  const gradeRecords = await prisma.gradeMaster.findMany();
  const gradeMap = new Map(gradeRecords.map((g) => [g.name, g.id]));

  const cbseRecord = await prisma.boardMaster.findUnique({ where: { name: "CBSE" } });
  const cbseId = cbseRecord?.id ?? null;

  let subjectCount = 0;

  // Insert common subjects (boardId = null)
  for (const [gradeName, subjects] of Object.entries(commonByGrade)) {
    const gradeId = gradeMap.get(gradeName);
    if (!gradeId) continue;
    for (const s of subjects) {
      await prisma.subjectMaster.create({
        data: {
          name: s.name,
          gradeId,
          boardId: null,
          icon: icon(s.name),
          sortOrder: s.sortOrder,
        },
      });
      subjectCount++;
    }
  }

  // Insert CBSE NCERT subjects (boardId = cbseId)
  if (cbseId) {
    for (const [gradeName, subjects] of Object.entries(cbseByGrade)) {
      const gradeId = gradeMap.get(gradeName);
      if (!gradeId) continue;
      for (const s of subjects) {
        await prisma.subjectMaster.create({
          data: {
            name: s.name,
            gradeId,
            boardId: cbseId,
            icon: icon(s.name),
            sortOrder: s.sortOrder,
          },
        });
        subjectCount++;
      }
    }
  }

  console.log(`✓ Seeded ${subjectCount} subjects (common + CBSE NCERT) across ${grades.length} grades`);

  // 4. Books + Chapters — delete all first then recreate
  await prisma.bookMaster.deleteMany();

  const subjectRecords = await prisma.subjectMaster.findMany({
    select: { id: true, name: true, gradeId: true, boardId: true },
  });

  // Helper: find subject id by name + grade + boardId
  function findSubject(name: string, gradeName: string, boardId: string | null): string | null {
    const gradeId = gradeMap.get(gradeName);
    if (!gradeId) return null;
    return subjectRecords.find(
      (s) => s.name === name && s.gradeId === gradeId && s.boardId === boardId
    )?.id ?? null;
  }

  type ChapterDef = { name: string; sortOrder: number; part?: number };
  type BookDef = { name: string; sortOrder: number; totalParts?: number; chapters: ChapterDef[] };

  async function seedBooks(subjectId: string, books: BookDef[]) {
    for (const b of books) {
      const book = await prisma.bookMaster.create({
        data: { name: b.name, subjectId, sortOrder: b.sortOrder, totalParts: b.totalParts ?? null },
      });
      for (const c of b.chapters) {
        await prisma.chapterMaster.create({
          data: { name: c.name, bookId: book.id, sortOrder: c.sortOrder, part: c.part ?? null },
        });
      }
    }
  }

  // ── CBSE Books & Chapters ───────────────────────────────────────────────────

  const cbseBooks: Array<{ subjectName: string; grade: string; books: BookDef[] }> = [
    // ── Grade 9 ──
    {
      subjectName: "Mathematics", grade: "9th",
      books: [{ name: "Mathematics Class 9", sortOrder: 1, chapters: [
        { name: "Number Systems", sortOrder: 1 },
        { name: "Polynomials", sortOrder: 2 },
        { name: "Coordinate Geometry", sortOrder: 3 },
        { name: "Linear Equations in Two Variables", sortOrder: 4 },
        { name: "Introduction to Euclid's Geometry", sortOrder: 5 },
        { name: "Lines and Angles", sortOrder: 6 },
        { name: "Triangles", sortOrder: 7 },
        { name: "Quadrilaterals", sortOrder: 8 },
        { name: "Circles", sortOrder: 9 },
        { name: "Heron's Formula", sortOrder: 10 },
        { name: "Surface Areas and Volumes", sortOrder: 11 },
        { name: "Statistics", sortOrder: 12 },
      ]}],
    },
    {
      subjectName: "Physics", grade: "9th",
      books: [{ name: "Science Class 9", sortOrder: 1, chapters: [
        { name: "Motion", sortOrder: 1 },
        { name: "Force and Laws of Motion", sortOrder: 2 },
        { name: "Gravitation", sortOrder: 3 },
        { name: "Work and Energy", sortOrder: 4 },
        { name: "Sound", sortOrder: 5 },
      ]}],
    },
    {
      subjectName: "Chemistry", grade: "9th",
      books: [{ name: "Science Class 9", sortOrder: 1, chapters: [
        { name: "Matter in Our Surroundings", sortOrder: 1 },
        { name: "Is Matter Around Us Pure", sortOrder: 2 },
        { name: "Atoms and Molecules", sortOrder: 3 },
        { name: "Structure of the Atom", sortOrder: 4 },
      ]}],
    },
    {
      subjectName: "Biology", grade: "9th",
      books: [{ name: "Science Class 9", sortOrder: 1, chapters: [
        { name: "The Fundamental Unit of Life", sortOrder: 1 },
        { name: "Tissues", sortOrder: 2 },
        { name: "Diversity in Living Organisms", sortOrder: 3 },
        { name: "Why Do We Fall Ill", sortOrder: 4 },
        { name: "Natural Resources", sortOrder: 5 },
        { name: "Improvement in Food Resources", sortOrder: 6 },
      ]}],
    },
    {
      subjectName: "English", grade: "9th",
      books: [
        { name: "Beehive", sortOrder: 1, chapters: [
          { name: "The Fun They Had", sortOrder: 1 },
          { name: "The Sound of Music", sortOrder: 2 },
          { name: "The Little Girl", sortOrder: 3 },
          { name: "A Truly Beautiful Mind", sortOrder: 4 },
          { name: "The Snake and the Mirror", sortOrder: 5 },
          { name: "My Childhood", sortOrder: 6 },
          { name: "Packing", sortOrder: 7 },
          { name: "Reach for the Top", sortOrder: 8 },
          { name: "The Bond of Love", sortOrder: 9 },
          { name: "Kathmandu", sortOrder: 10 },
          { name: "If I Were You", sortOrder: 11 },
        ]},
      ],
    },
    // ── Grade 10 ──
    {
      subjectName: "Mathematics", grade: "10th",
      books: [{ name: "Mathematics Class 10", sortOrder: 1, chapters: [
        { name: "Real Numbers", sortOrder: 1 },
        { name: "Polynomials", sortOrder: 2 },
        { name: "Pair of Linear Equations", sortOrder: 3 },
        { name: "Quadratic Equations", sortOrder: 4 },
        { name: "Arithmetic Progressions", sortOrder: 5 },
        { name: "Triangles", sortOrder: 6 },
        { name: "Coordinate Geometry", sortOrder: 7 },
        { name: "Introduction to Trigonometry", sortOrder: 8 },
        { name: "Applications of Trigonometry", sortOrder: 9 },
        { name: "Circles", sortOrder: 10 },
        { name: "Areas Related to Circles", sortOrder: 11 },
        { name: "Surface Areas and Volumes", sortOrder: 12 },
        { name: "Statistics", sortOrder: 13 },
        { name: "Probability", sortOrder: 14 },
      ]}],
    },
    {
      subjectName: "Physics", grade: "10th",
      books: [{ name: "Science Class 10", sortOrder: 1, chapters: [
        { name: "Light – Reflection and Refraction", sortOrder: 1 },
        { name: "Human Eye and the Colourful World", sortOrder: 2 },
        { name: "Electricity", sortOrder: 3 },
        { name: "Magnetic Effects of Electric Current", sortOrder: 4 },
        { name: "Sources of Energy", sortOrder: 5 },
      ]}],
    },
    {
      subjectName: "Chemistry", grade: "10th",
      books: [{ name: "Science Class 10", sortOrder: 1, chapters: [
        { name: "Chemical Reactions and Equations", sortOrder: 1 },
        { name: "Acids, Bases and Salts", sortOrder: 2 },
        { name: "Metals and Non-metals", sortOrder: 3 },
        { name: "Carbon and its Compounds", sortOrder: 4 },
        { name: "Periodic Classification of Elements", sortOrder: 5 },
      ]}],
    },
    {
      subjectName: "Biology", grade: "10th",
      books: [{ name: "Science Class 10", sortOrder: 1, chapters: [
        { name: "Life Processes", sortOrder: 1 },
        { name: "Control and Coordination", sortOrder: 2 },
        { name: "How do Organisms Reproduce", sortOrder: 3 },
        { name: "Heredity and Evolution", sortOrder: 4 },
        { name: "Our Environment", sortOrder: 5 },
        { name: "Management of Natural Resources", sortOrder: 6 },
      ]}],
    },
    // ── Grade 11 ──
    {
      subjectName: "Physics", grade: "11th",
      books: [
        { name: "Physics Class 11", sortOrder: 1, totalParts: 2, chapters: [
          { name: "Physical World", sortOrder: 1, part: 1 },
          { name: "Units and Measurements", sortOrder: 2, part: 1 },
          { name: "Motion in a Straight Line", sortOrder: 3, part: 1 },
          { name: "Motion in a Plane", sortOrder: 4, part: 1 },
          { name: "Laws of Motion", sortOrder: 5, part: 1 },
          { name: "Work, Energy and Power", sortOrder: 6, part: 1 },
          { name: "System of Particles and Rotational Motion", sortOrder: 7, part: 1 },
          { name: "Gravitation", sortOrder: 8, part: 1 },
          { name: "Mechanical Properties of Solids", sortOrder: 9, part: 2 },
          { name: "Mechanical Properties of Fluids", sortOrder: 10, part: 2 },
          { name: "Thermal Properties of Matter", sortOrder: 11, part: 2 },
          { name: "Thermodynamics", sortOrder: 12, part: 2 },
          { name: "Kinetic Theory", sortOrder: 13, part: 2 },
          { name: "Oscillations", sortOrder: 14, part: 2 },
          { name: "Waves", sortOrder: 15, part: 2 },
        ]},
      ],
    },
    {
      subjectName: "Chemistry", grade: "11th",
      books: [
        { name: "Chemistry Class 11", sortOrder: 1, totalParts: 2, chapters: [
          { name: "Some Basic Concepts of Chemistry", sortOrder: 1, part: 1 },
          { name: "Structure of Atom", sortOrder: 2, part: 1 },
          { name: "Classification of Elements and Periodicity", sortOrder: 3, part: 1 },
          { name: "Chemical Bonding and Molecular Structure", sortOrder: 4, part: 1 },
          { name: "Thermodynamics", sortOrder: 5, part: 1 },
          { name: "Equilibrium", sortOrder: 6, part: 1 },
          { name: "Redox Reactions", sortOrder: 7, part: 2 },
          { name: "Organic Chemistry: Basic Principles", sortOrder: 8, part: 2 },
          { name: "Hydrocarbons", sortOrder: 9, part: 2 },
          { name: "Environmental Chemistry", sortOrder: 10, part: 2 },
        ]},
      ],
    },
    {
      subjectName: "Biology", grade: "11th",
      books: [{ name: "Biology Class 11", sortOrder: 1, chapters: [
        { name: "The Living World", sortOrder: 1 },
        { name: "Biological Classification", sortOrder: 2 },
        { name: "Plant Kingdom", sortOrder: 3 },
        { name: "Animal Kingdom", sortOrder: 4 },
        { name: "Morphology of Flowering Plants", sortOrder: 5 },
        { name: "Anatomy of Flowering Plants", sortOrder: 6 },
        { name: "Structural Organisation in Animals", sortOrder: 7 },
        { name: "Cell: The Unit of Life", sortOrder: 8 },
        { name: "Biomolecules", sortOrder: 9 },
        { name: "Cell Cycle and Cell Division", sortOrder: 10 },
        { name: "Photosynthesis in Higher Plants", sortOrder: 11 },
        { name: "Respiration in Plants", sortOrder: 12 },
        { name: "Plant Growth and Development", sortOrder: 13 },
        { name: "Breathing and Exchange of Gases", sortOrder: 14 },
        { name: "Body Fluids and Circulation", sortOrder: 15 },
        { name: "Excretory Products and their Elimination", sortOrder: 16 },
        { name: "Locomotion and Movement", sortOrder: 17 },
        { name: "Neural Control and Coordination", sortOrder: 18 },
        { name: "Chemical Coordination and Integration", sortOrder: 19 },
      ]}],
    },
    {
      subjectName: "Mathematics", grade: "11th",
      books: [{ name: "Mathematics Class 11", sortOrder: 1, chapters: [
        { name: "Sets", sortOrder: 1 },
        { name: "Relations and Functions", sortOrder: 2 },
        { name: "Trigonometric Functions", sortOrder: 3 },
        { name: "Complex Numbers and Quadratic Equations", sortOrder: 4 },
        { name: "Linear Inequalities", sortOrder: 5 },
        { name: "Permutations and Combinations", sortOrder: 6 },
        { name: "Binomial Theorem", sortOrder: 7 },
        { name: "Sequences and Series", sortOrder: 8 },
        { name: "Straight Lines", sortOrder: 9 },
        { name: "Conic Sections", sortOrder: 10 },
        { name: "Introduction to Three Dimensional Geometry", sortOrder: 11 },
        { name: "Limits and Derivatives", sortOrder: 12 },
        { name: "Statistics", sortOrder: 13 },
        { name: "Probability", sortOrder: 14 },
      ]}],
    },
    // ── Grade 12 ──
    {
      subjectName: "Physics", grade: "12th",
      books: [
        { name: "Physics Class 12", sortOrder: 1, totalParts: 2, chapters: [
          { name: "Electric Charges and Fields", sortOrder: 1, part: 1 },
          { name: "Electrostatic Potential and Capacitance", sortOrder: 2, part: 1 },
          { name: "Current Electricity", sortOrder: 3, part: 1 },
          { name: "Moving Charges and Magnetism", sortOrder: 4, part: 1 },
          { name: "Magnetism and Matter", sortOrder: 5, part: 1 },
          { name: "Electromagnetic Induction", sortOrder: 6, part: 1 },
          { name: "Alternating Current", sortOrder: 7, part: 1 },
          { name: "Electromagnetic Waves", sortOrder: 8, part: 1 },
          { name: "Ray Optics and Optical Instruments", sortOrder: 9, part: 2 },
          { name: "Wave Optics", sortOrder: 10, part: 2 },
          { name: "Dual Nature of Radiation and Matter", sortOrder: 11, part: 2 },
          { name: "Atoms", sortOrder: 12, part: 2 },
          { name: "Nuclei", sortOrder: 13, part: 2 },
          { name: "Semiconductor Electronics", sortOrder: 14, part: 2 },
        ]},
      ],
    },
    {
      subjectName: "Chemistry", grade: "12th",
      books: [
        { name: "Chemistry Class 12", sortOrder: 1, totalParts: 2, chapters: [
          { name: "The Solid State", sortOrder: 1, part: 1 },
          { name: "Solutions", sortOrder: 2, part: 1 },
          { name: "Electrochemistry", sortOrder: 3, part: 1 },
          { name: "Chemical Kinetics", sortOrder: 4, part: 1 },
          { name: "Surface Chemistry", sortOrder: 5, part: 1 },
          { name: "General Principles and Processes of Isolation of Elements", sortOrder: 6, part: 1 },
          { name: "The p-Block Elements", sortOrder: 7, part: 1 },
          { name: "The d and f Block Elements", sortOrder: 8, part: 1 },
          { name: "Coordination Compounds", sortOrder: 9, part: 1 },
          { name: "Haloalkanes and Haloarenes", sortOrder: 10, part: 2 },
          { name: "Alcohols, Phenols and Ethers", sortOrder: 11, part: 2 },
          { name: "Aldehydes, Ketones and Carboxylic Acids", sortOrder: 12, part: 2 },
          { name: "Amines", sortOrder: 13, part: 2 },
          { name: "Biomolecules", sortOrder: 14, part: 2 },
          { name: "Polymers", sortOrder: 15, part: 2 },
          { name: "Chemistry in Everyday Life", sortOrder: 16, part: 2 },
        ]},
      ],
    },
    {
      subjectName: "Biology", grade: "12th",
      books: [{ name: "Biology Class 12", sortOrder: 1, chapters: [
        { name: "Sexual Reproduction in Flowering Plants", sortOrder: 1 },
        { name: "Human Reproduction", sortOrder: 2 },
        { name: "Reproductive Health", sortOrder: 3 },
        { name: "Principles of Inheritance and Variation", sortOrder: 4 },
        { name: "Molecular Basis of Inheritance", sortOrder: 5 },
        { name: "Evolution", sortOrder: 6 },
        { name: "Human Health and Disease", sortOrder: 7 },
        { name: "Microbes in Human Welfare", sortOrder: 8 },
        { name: "Biotechnology: Principles and Processes", sortOrder: 9 },
        { name: "Biotechnology and its Applications", sortOrder: 10 },
        { name: "Organisms and Populations", sortOrder: 11 },
        { name: "Ecosystem", sortOrder: 12 },
        { name: "Biodiversity and Conservation", sortOrder: 13 },
        { name: "Environmental Issues", sortOrder: 14 },
      ]}],
    },
    {
      subjectName: "Mathematics", grade: "12th",
      books: [{ name: "Mathematics Class 12", sortOrder: 1, chapters: [
        { name: "Relations and Functions", sortOrder: 1 },
        { name: "Inverse Trigonometric Functions", sortOrder: 2 },
        { name: "Matrices", sortOrder: 3 },
        { name: "Determinants", sortOrder: 4 },
        { name: "Continuity and Differentiability", sortOrder: 5 },
        { name: "Application of Derivatives", sortOrder: 6 },
        { name: "Integrals", sortOrder: 7 },
        { name: "Application of Integrals", sortOrder: 8 },
        { name: "Differential Equations", sortOrder: 9 },
        { name: "Vector Algebra", sortOrder: 10 },
        { name: "Three Dimensional Geometry", sortOrder: 11 },
        { name: "Linear Programming", sortOrder: 12 },
        { name: "Probability", sortOrder: 13 },
      ]}],
    },
  ];

  let bookCount = 0;
  let chapterCount = 0;

  for (const entry of cbseBooks) {
    const subjectId = findSubject(entry.subjectName, entry.grade, cbseId);
    if (!subjectId) continue;
    for (const b of entry.books) {
      const book = await prisma.bookMaster.create({
        data: { name: b.name, subjectId, sortOrder: b.sortOrder, totalParts: b.totalParts ?? null },
      });
      bookCount++;
      for (const c of b.chapters) {
        await prisma.chapterMaster.create({
          data: { name: c.name, bookId: book.id, sortOrder: c.sortOrder, part: c.part ?? null },
        });
        chapterCount++;
      }
    }
  }

  // ── Common subjects: simple numbered chapters ───────────────────────────────
  const commonBookSeeds: Array<{ subjectName: string; grade: string; bookName: string; numChapters: number }> = [
    { subjectName: "Mathematics", grade: "1st",  bookName: "Math Magic 1",   numChapters: 13 },
    { subjectName: "Mathematics", grade: "2nd",  bookName: "Math Magic 2",   numChapters: 15 },
    { subjectName: "Mathematics", grade: "3rd",  bookName: "Math Magic 3",   numChapters: 14 },
    { subjectName: "Mathematics", grade: "4th",  bookName: "Math Magic 4",   numChapters: 14 },
    { subjectName: "Mathematics", grade: "5th",  bookName: "Math Magic 5",   numChapters: 14 },
    { subjectName: "Mathematics", grade: "6th",  bookName: "Mathematics 6",  numChapters: 14 },
    { subjectName: "Mathematics", grade: "7th",  bookName: "Mathematics 7",  numChapters: 15 },
    { subjectName: "Mathematics", grade: "8th",  bookName: "Mathematics 8",  numChapters: 16 },
    { subjectName: "Science",     grade: "3rd",  bookName: "Science 3",      numChapters: 10 },
    { subjectName: "Science",     grade: "4th",  bookName: "Science 4",      numChapters: 10 },
    { subjectName: "Science",     grade: "5th",  bookName: "Science 5",      numChapters: 10 },
    { subjectName: "Science",     grade: "6th",  bookName: "Science 6",      numChapters: 16 },
    { subjectName: "Science",     grade: "7th",  bookName: "Science 7",      numChapters: 18 },
    { subjectName: "Science",     grade: "8th",  bookName: "Science 8",      numChapters: 18 },
    { subjectName: "English",     grade: "1st",  bookName: "Marigold 1",     numChapters: 10 },
    { subjectName: "English",     grade: "2nd",  bookName: "Marigold 2",     numChapters: 10 },
    { subjectName: "English",     grade: "3rd",  bookName: "Marigold 3",     numChapters: 10 },
    { subjectName: "English",     grade: "4th",  bookName: "Marigold 4",     numChapters: 10 },
    { subjectName: "English",     grade: "5th",  bookName: "Marigold 5",     numChapters: 10 },
    { subjectName: "English",     grade: "6th",  bookName: "Honeysuckle",    numChapters: 10 },
    { subjectName: "English",     grade: "7th",  bookName: "Honeycomb",      numChapters: 10 },
    { subjectName: "English",     grade: "8th",  bookName: "Honeydew",       numChapters: 10 },
  ];

  for (const entry of commonBookSeeds) {
    const subjectId = findSubject(entry.subjectName, entry.grade, null);
    if (!subjectId) continue;
    const book = await prisma.bookMaster.create({
      data: { name: entry.bookName, subjectId, sortOrder: 1 },
    });
    bookCount++;
    for (let i = 1; i <= entry.numChapters; i++) {
      await prisma.chapterMaster.create({
        data: { name: `Chapter ${i}`, bookId: book.id, sortOrder: i },
      });
      chapterCount++;
    }
  }

  console.log(`✓ Seeded ${bookCount} books and ${chapterCount} chapters`);
}

main()
  .catch(console.error)
  .finally(() => prisma.$disconnect());
