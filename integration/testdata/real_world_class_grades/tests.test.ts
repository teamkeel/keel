import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - class enrollment counts and grades", async () => {
  // Create teacher and class
  const teacher = await models.teacher.create({ name: "Dr. Smith" });
  const class1 = await models.class.create({
    name: "Math 101",
    teacherId: teacher.id,
  });

  // Initially, class should have no enrollments
  let classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(0);
  expect(classData?.averageGrade).toBe(0);
  expect(classData?.medianGrade).toBe(0);
  expect(classData?.lowestGrade).toBe(0);
  expect(classData?.highestGrade).toBe(0);

  // Create active students
  const student1 = await models.student.create({
    name: "Alice",
    isActive: true,
  });
  const student2 = await models.student.create({ name: "Bob", isActive: true });
  const student3 = await models.student.create({
    name: "Charlie",
    isActive: true,
  });
  const inactiveStudent = await models.student.create({
    name: "David",
    isActive: false,
  });

  // Enroll active students with grades
  await models.enrollment.create({
    studentId: student1.id,
    classId: class1.id,
    grade: 85,
  });
  await models.enrollment.create({
    studentId: student2.id,
    classId: class1.id,
    grade: 92,
  });
  await models.enrollment.create({
    studentId: student3.id,
    classId: class1.id,
    grade: 78,
  });

  // Enroll inactive student (should not be counted in computed fields)
  await models.enrollment.create({
    studentId: inactiveStudent.id,
    classId: class1.id,
    grade: 95,
  });

  // Check computed fields - only active students should be counted
  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3);
  expect(classData?.averageGrade).toBe(85); // (85 + 92 + 78) / 3 = 85
  expect(classData?.medianGrade).toBe(85); // middle value when sorted: 78, 85, 92
  expect(classData?.lowestGrade).toBe(78);
  expect(classData?.highestGrade).toBe(92);

  // Update a grade and verify computed fields update
  await models.enrollment.update(
    {
      id: (await models.enrollment.findOne({
        studentId: student1.id,
        classId: class1.id,
      }))!.id,
    },
    { grade: 95 }
  );

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3);
  expect(classData?.averageGrade).toBe((95 + 92 + 78) / 3); // (95 + 92 + 78) / 3 = 88.33
  expect(classData?.medianGrade).toBe(92); // middle value when sorted: 78, 92, 95
  expect(classData?.lowestGrade).toBe(78);
  expect(classData?.highestGrade).toBe(95);

  // Activate inactive student and verify computed fields update
  await models.student.update({ id: inactiveStudent.id }, { isActive: true });

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(4);
  expect(classData?.averageGrade).toBe((95 + 92 + 78 + 95) / 4); // (95 + 92 + 78 + 95) / 4 = 87.5
  expect(classData?.medianGrade).toBe(93.5); // middle value when sorted: 78, 92, 95, 95
  expect(classData?.lowestGrade).toBe(78);
  expect(classData?.highestGrade).toBe(95);

  // Deactivate a student and verify computed fields update
  await models.student.update({ id: student2.id }, { isActive: false });

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3);
  expect(classData?.averageGrade).toBe((95 + 78 + 95) / 3); // (95 + 78 + 95) / 3 = 89.33
  expect(classData?.medianGrade).toBe(95); // middle value when sorted: 78, 95, 95
  expect(classData?.lowestGrade).toBe(78);
  expect(classData?.highestGrade).toBe(95);
});

test("computed fields - class with null grades", async () => {
  const teacher = await models.teacher.create({ name: "Dr. Johnson" });
  const class1 = await models.class.create({
    name: "Physics 101",
    teacherId: teacher.id,
  });

  const student1 = await models.student.create({ name: "Eve", isActive: true });
  const student2 = await models.student.create({
    name: "Frank",
    isActive: true,
  });
  const student3 = await models.student.create({
    name: "Grace",
    isActive: true,
  });

  // Enroll students with some null grades
  await models.enrollment.create({
    studentId: student1.id,
    classId: class1.id,
    grade: 88,
  });
  await models.enrollment.create({
    studentId: student2.id,
    classId: class1.id,
    grade: null,
  });
  await models.enrollment.create({
    studentId: student3.id,
    classId: class1.id,
    grade: 76,
  });

  // Check computed fields - null grades should be ignored
  const classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3); // All students are active
  expect(classData?.averageGrade).toBe(82); // (88 + 76) / 2 = 82 (null grade ignored)
  expect(classData?.medianGrade).toBe(82); // middle value when sorted: 76, 88
  expect(classData?.lowestGrade).toBe(76);
  expect(classData?.highestGrade).toBe(88);

  // Add a grade to the null enrollment
  await models.enrollment.update(
    {
      id: (await models.enrollment.findOne({
        studentId: student2.id,
        classId: class1.id,
      }))!.id,
    },
    { grade: 90 }
  );

  const updatedClassData = await models.class.findOne({ id: class1.id });
  expect(updatedClassData?.numberOfEnrollments).toBe(3);
  expect(updatedClassData?.averageGrade).toBe((88 + 90 + 76) / 3); // (88 + 90 + 76) / 3 = 84.67
  expect(updatedClassData?.medianGrade).toBe(88); // middle value when sorted: 76, 88, 90
  expect(updatedClassData?.lowestGrade).toBe(76);
  expect(updatedClassData?.highestGrade).toBe(90);
});

test("computed fields - teacher grade average across multiple classes", async () => {
  const teacher = await models.teacher.create({ name: "Prof. Brown" });

  // Create multiple classes
  const mathClass = await models.class.create({
    name: "Advanced Math",
    teacherId: teacher.id,
  });
  const scienceClass = await models.class.create({
    name: "Advanced Science",
    teacherId: teacher.id,
  });

  // Create students
  const student1 = await models.student.create({
    name: "Hannah",
    isActive: true,
  });
  const student2 = await models.student.create({ name: "Ian", isActive: true });
  const student3 = await models.student.create({
    name: "Julia",
    isActive: true,
  });
  const inactiveStudent = await models.student.create({
    name: "Kevin",
    isActive: false,
  });

  // Enroll students in math class
  await models.enrollment.create({
    studentId: student1.id,
    classId: mathClass.id,
    grade: 85,
  });
  await models.enrollment.create({
    studentId: student2.id,
    classId: mathClass.id,
    grade: 92,
  });
  await models.enrollment.create({
    studentId: inactiveStudent.id,
    classId: mathClass.id,
    grade: 70,
  });

  // Enroll students in science class
  await models.enrollment.create({
    studentId: student2.id,
    classId: scienceClass.id,
    grade: 88,
  });
  await models.enrollment.create({
    studentId: student3.id,
    classId: scienceClass.id,
    grade: 95,
  });
  await models.enrollment.create({
    studentId: inactiveStudent.id,
    classId: scienceClass.id,
    grade: 65,
  });

  // Check teacher's grade average - should only include active students
  const teacherData = await models.teacher.findOne({ id: teacher.id });
  expect(teacherData?.gradeAverage).toBe(90); // (85 + 92 + 88 + 95) / 4 = 90

  // Deactivate a student and verify teacher average updates
  await models.student.update({ id: student1.id }, { isActive: false });

  const updatedTeacherData = await models.teacher.findOne({ id: teacher.id });
  expect(updatedTeacherData?.gradeAverage).toBe((92 + 88 + 95) / 3); // (92 + 88 + 95) / 3 = 91.67

  // Add a new class with grades
  const historyClass = await models.class.create({
    name: "History",
    teacherId: teacher.id,
  });
  const newStudent = await models.student.create({
    name: "Liam",
    isActive: true,
  });
  await models.enrollment.create({
    studentId: newStudent.id,
    classId: historyClass.id,
    grade: 87,
  });

  const finalTeacherData = await models.teacher.findOne({ id: teacher.id });
  expect(finalTeacherData?.gradeAverage).toBe(90.5); // (92 + 88 + 95 + 87) / 4 = 90.5
});

test("computed fields - student result average", async () => {
  const teacher1 = await models.teacher.create({ name: "Dr. Wilson" });
  const teacher2 = await models.teacher.create({ name: "Dr. Davis" });

  const class1 = await models.class.create({
    name: "English",
    teacherId: teacher1.id,
  });
  const class2 = await models.class.create({
    name: "History",
    teacherId: teacher2.id,
  });

  const student = await models.student.create({ name: "Maya", isActive: true });

  // Initially, student should have no result
  let studentData = await models.student.findOne({ id: student.id });
  expect(studentData?.result).toBe(0);

  // Enroll in first class
  await models.enrollment.create({
    studentId: student.id,
    classId: class1.id,
    grade: 85,
  });

  studentData = await models.student.findOne({ id: student.id });
  expect(studentData?.result).toBe(85);

  // Enroll in second class
  await models.enrollment.create({
    studentId: student.id,
    classId: class2.id,
    grade: 92,
  });

  studentData = await models.student.findOne({ id: student.id });
  expect(studentData?.result).toBe(88.5); // (85 + 92) / 2 = 88.5

  // Update a grade
  await models.enrollment.update(
    {
      id: (await models.enrollment.findOne({
        studentId: student.id,
        classId: class1.id,
      }))!.id,
    },
    { grade: 90 }
  );

  studentData = await models.student.findOne({ id: student.id });
  expect(studentData?.result).toBe(91); // (90 + 92) / 2 = 91

  // Add a null grade (should be ignored)
  const class3 = await models.class.create({
    name: "Art",
    teacherId: teacher1.id,
  });
  await models.enrollment.create({
    studentId: student.id,
    classId: class3.id,
    grade: null,
  });

  studentData = await models.student.findOne({ id: student.id });
  expect(studentData?.result).toBe(91); // (90 + 92) / 2 = 91 (null grade ignored)
});

test("computed fields - edge cases with no active students", async () => {
  const teacher = await models.teacher.create({ name: "Dr. Thompson" });
  const class1 = await models.class.create({
    name: "Chemistry",
    teacherId: teacher.id,
  });

  // Create only inactive students
  const inactiveStudent1 = await models.student.create({
    name: "Noah",
    isActive: false,
  });
  const inactiveStudent2 = await models.student.create({
    name: "Olivia",
    isActive: false,
  });

  await models.enrollment.create({
    studentId: inactiveStudent1.id,
    classId: class1.id,
    grade: 85,
  });
  await models.enrollment.create({
    studentId: inactiveStudent2.id,
    classId: class1.id,
    grade: 92,
  });

  // Check computed fields - should be 0/null since no active students
  const classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(0);
  expect(classData?.averageGrade).toBe(0);
  expect(classData?.medianGrade).toBe(0);
  expect(classData?.lowestGrade).toBe(0);
  expect(classData?.highestGrade).toBe(0);

  // Teacher should also have null average
  const teacherData = await models.teacher.findOne({ id: teacher.id });
  expect(teacherData?.gradeAverage).toBe(0);
});

test("computed fields - single student scenarios", async () => {
  const teacher = await models.teacher.create({ name: "Dr. Garcia" });
  const class1 = await models.class.create({
    name: "Philosophy",
    teacherId: teacher.id,
  });
  const student = await models.student.create({
    name: "Peter",
    isActive: true,
  });

  // Single student enrollment
  await models.enrollment.create({
    studentId: student.id,
    classId: class1.id,
    grade: 88,
  });

  const classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(1);
  expect(classData?.averageGrade).toBe(88);
  expect(classData?.medianGrade).toBe(88); // Single value is its own median
  expect(classData?.lowestGrade).toBe(88);
  expect(classData?.highestGrade).toBe(88);

  // Teacher average should match
  const teacherData = await models.teacher.findOne({ id: teacher.id });
  expect(teacherData?.gradeAverage).toBe(88);
});

test("computed fields - complex scenario with multiple updates", async () => {
  const teacher = await models.teacher.create({ name: "Dr. Martinez" });
  const class1 = await models.class.create({
    name: "Computer Science",
    teacherId: teacher.id,
  });

  // Create students with different active states
  const student1 = await models.student.create({
    name: "Quinn",
    isActive: true,
  });
  const student2 = await models.student.create({
    name: "Rachel",
    isActive: true,
  });
  const student3 = await models.student.create({
    name: "Sam",
    isActive: false,
  });
  const student4 = await models.student.create({
    name: "Tina",
    isActive: true,
  });

  // Initial enrollments
  await models.enrollment.create({
    studentId: student1.id,
    classId: class1.id,
    grade: 75,
  });
  await models.enrollment.create({
    studentId: student2.id,
    classId: class1.id,
    grade: 85,
  });
  await models.enrollment.create({
    studentId: student3.id,
    classId: class1.id,
    grade: 95,
  });
  await models.enrollment.create({
    studentId: student4.id,
    classId: class1.id,
    grade: 90,
  });

  // Check initial state - only active students counted
  let classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3);
  expect(classData?.averageGrade).toBe((75 + 85 + 90) / 3); // (75 + 85 + 90) / 3 = 83.33
  expect(classData?.medianGrade).toBe(85); // middle value: 75, 85, 90
  expect(classData?.lowestGrade).toBe(75);
  expect(classData?.highestGrade).toBe(90);

  // Activate inactive student
  await models.student.update({ id: student3.id }, { isActive: true });

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(4);
  expect(classData?.averageGrade).toBe(86.25); // (75 + 85 + 95 + 90) / 4 = 86.25
  expect(classData?.medianGrade).toBe(87.5); // middle value: 75, 85, 90, 95
  expect(classData?.lowestGrade).toBe(75);
  expect(classData?.highestGrade).toBe(95);

  // Update grades
  await models.enrollment.update(
    {
      id: (await models.enrollment.findOne({
        studentId: student1.id,
        classId: class1.id,
      }))!.id,
    },
    { grade: 80 }
  );
  await models.enrollment.update(
    {
      id: (await models.enrollment.findOne({
        studentId: student2.id,
        classId: class1.id,
      }))!.id,
    },
    { grade: 88 }
  );

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(4);
  expect(classData?.averageGrade).toBe(88.25); // (80 + 88 + 95 + 90) / 4 = 88.25
  expect(classData?.medianGrade).toBe(89); // middle value: 80, 88, 90, 95
  expect(classData?.lowestGrade).toBe(80);
  expect(classData?.highestGrade).toBe(95);

  // Deactivate a student
  await models.student.update({ id: student4.id }, { isActive: false });

  classData = await models.class.findOne({ id: class1.id });
  expect(classData?.numberOfEnrollments).toBe(3);
  expect(classData?.averageGrade).toBe((80 + 88 + 95) / 3); // (80 + 88 + 95) / 3 = 87.67
  expect(classData?.medianGrade).toBe(88); // middle value: 80, 88, 95
  expect(classData?.lowestGrade).toBe(80);
  expect(classData?.highestGrade).toBe(95);
});
